package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

// --- Invoice ---

type InvoiceRepo struct{ pool database.Pool }

func NewInvoiceRepo(pool database.Pool) *InvoiceRepo { return &InvoiceRepo{pool: pool} }

type InvoiceInput struct {
	Date                 string
	DueDate              *string
	PONumber             *string
	BuyerName            *string
	BuyerAddress         *string
	BuyerContact         *string
	Discount             decimal.Decimal
	TaxRate              decimal.Decimal
	PaymentBank          *string
	PaymentAccountNumber *string
	PaymentAccountName   *string
	Notes                *string
	FinanceName          *string
	Total                decimal.Decimal
	Status               models.InvoiceStatus
	Items                []models.InvoiceItem
}

const invoiceCols = `id, "projectId", number, date, "dueDate", "poNumber", "buyerName", "buyerAddress", "buyerContact",
	discount::text, "taxRate"::text, "paymentBank", "paymentAccountNumber", "paymentAccountName", notes, "financeName",
	total::text, status, "createdAt", "updatedAt"`

func scanInvoice(row pgx.Row) (*models.Invoice, error) {
	var inv models.Invoice
	var total, discount, taxRate string
	var dueDate, poNumber, buyerName, buyerAddress, buyerContact, paymentBank, paymentAccNo, paymentAccName, notes, financeName sql.NullString
	if err := row.Scan(&inv.ID, &inv.ProjectID, &inv.Number, &inv.Date, &dueDate, &poNumber, &buyerName, &buyerAddress, &buyerContact,
		&discount, &taxRate, &paymentBank, &paymentAccNo, &paymentAccName, &notes, &financeName,
		&total, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
		return nil, err
	}
	inv.Discount = scanDec(discount)
	inv.TaxRate = scanDec(taxRate)
	inv.Total = scanDec(total)
	inv.DueDate = timePtr(dueDate)
	inv.PONumber = strPtr(poNumber)
	inv.BuyerName = strPtr(buyerName)
	inv.BuyerAddress = strPtr(buyerAddress)
	inv.BuyerContact = strPtr(buyerContact)
	inv.PaymentBank = strPtr(paymentBank)
	inv.PaymentAccountNumber = strPtr(paymentAccNo)
	inv.PaymentAccountName = strPtr(paymentAccName)
	inv.Notes = strPtr(notes)
	inv.FinanceName = strPtr(financeName)
	return &inv, nil
}

func (r *InvoiceRepo) loadItems(ctx context.Context, invs []*models.Invoice) error {
	if len(invs) == 0 {
		return nil
	}
	ids := make([]int32, 0, len(invs))
	byID := make(map[int32]*models.Invoice, len(invs))
	for _, inv := range invs {
		ids = append(ids, inv.ID)
		byID[inv.ID] = inv
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, "invoiceId", description, qty::text, unit, "unitPrice"::text
		FROM invoice_items WHERE "invoiceId" = ANY($1) ORDER BY id ASC`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.InvoiceItem
		var qty, unitPrice string
		if err := rows.Scan(&item.ID, &item.InvoiceID, &item.Description, &qty, &item.Unit, &unitPrice); err != nil {
			return err
		}
		item.Qty = scanDec(qty)
		item.UnitPrice = scanDec(unitPrice)
		if inv, ok := byID[item.InvoiceID]; ok {
			inv.Items = append(inv.Items, item)
		}
	}
	return rows.Err()
}

func (r *InvoiceRepo) List(ctx context.Context, projectID int32) ([]models.Invoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+invoiceCols+` FROM invoices WHERE "projectId" = $1 ORDER BY id DESC`, projectID)
	if err != nil {
		return nil, err
	}
	var out []*models.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, inv)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, out); err != nil {
		return nil, err
	}
	res := make([]models.Invoice, 0, len(out))
	for _, inv := range out {
		res = append(res, *inv)
	}
	return res, nil
}

func (r *InvoiceRepo) Get(ctx context.Context, projectID, id int32) (*models.Invoice, error) {
	inv, err := scanInvoice(r.pool.QueryRow(ctx, `
		SELECT `+invoiceCols+` FROM invoices WHERE id = $1 AND "projectId" = $2`, id, projectID))
	if err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, []*models.Invoice{inv}); err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InvoiceRepo) insertItems(ctx context.Context, tx pgx.Tx, invoiceID int32, items []models.InvoiceItem) error {
	if len(items) == 0 {
		return nil
	}
	for _, it := range items {
		if _, err := tx.Exec(ctx, `
			INSERT INTO invoice_items ("invoiceId", description, qty, unit, "unitPrice")
			VALUES ($1,$2,$3,$4,$5)`,
			invoiceID, it.Description, it.Qty.String(), it.Unit, it.UnitPrice.String()); err != nil {
			return err
		}
	}
	return nil
}

func (r *InvoiceRepo) Create(ctx context.Context, projectID int32, in *InvoiceInput) (*models.Invoice, error) {
	var seq int
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX((substring(number from '[0-9]+$'))::int), 0) + 1 FROM invoices
		WHERE "projectId" = $1 AND to_char(date, 'YYYY-MM') = to_char($2::date, 'YYYY-MM')`,
		projectID, in.Date).Scan(&seq); err != nil {
		return nil, err
	}
	d, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, err
	}
	number := fmt.Sprintf("INV/%s/%s", d.Format("2006/01"), pad3(seq))
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	inv, err := scanInvoice(tx.QueryRow(ctx, `
		INSERT INTO invoices ("projectId", number, date, "dueDate", "poNumber", "buyerName", "buyerAddress", "buyerContact",
			discount, "taxRate", "paymentBank", "paymentAccountNumber", "paymentAccountName", notes, "financeName", total, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING `+invoiceCols,
		projectID, number, in.Date, in.DueDate, in.PONumber, in.BuyerName, in.BuyerAddress, in.BuyerContact,
		in.Discount.String(), in.TaxRate.String(), in.PaymentBank, in.PaymentAccountNumber, in.PaymentAccountName, in.Notes, in.FinanceName,
		in.Total.String(), in.Status))
	if err != nil {
		return nil, err
	}
	if err := r.insertItems(ctx, tx, inv.ID, in.Items); err != nil {
		return nil, err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}
	inv.Items = in.Items
	return inv, nil
}

// Update overwrites the full document (including items) on a draft invoice.
func (r *InvoiceRepo) Update(ctx context.Context, projectID, id int32, in *InvoiceInput) (*models.Invoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	inv, err := scanInvoice(tx.QueryRow(ctx, `
		UPDATE invoices SET
			date = $3,
			"dueDate" = $4,
			"poNumber" = $5,
			"buyerName" = $6,
			"buyerAddress" = $7,
			"buyerContact" = $8,
			discount = $9,
			"taxRate" = $10,
			"paymentBank" = $11,
			"paymentAccountNumber" = $12,
			"paymentAccountName" = $13,
			notes = $14,
			"financeName" = $15,
			total = $16,
			"updatedAt" = NOW()
		WHERE id = $1 AND "projectId" = $2
		RETURNING `+invoiceCols,
		id, projectID, in.Date, in.DueDate, in.PONumber, in.BuyerName, in.BuyerAddress, in.BuyerContact,
		in.Discount.String(), in.TaxRate.String(), in.PaymentBank, in.PaymentAccountNumber, in.PaymentAccountName, in.Notes, in.FinanceName,
		in.Total.String()))
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM invoice_items WHERE "invoiceId" = $1`, id); err != nil {
		return nil, err
	}
	if err := r.insertItems(ctx, tx, id, in.Items); err != nil {
		return nil, err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, []*models.Invoice{inv}); err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InvoiceRepo) UpdateStatus(ctx context.Context, projectID, id int32, status models.InvoiceStatus) (*models.Invoice, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE invoices SET status = $3, "updatedAt" = NOW()
		WHERE id = $1 AND "projectId" = $2
		RETURNING `+invoiceCols,
		id, projectID, status)
	return scanInvoice(row)
}

func pad3(n int) string {
	s := strconv.Itoa(n)
	for len(s) < 3 {
		s = "0" + s
	}
	return s
}

// --- Logistics ---

type LogisticsRepo struct{ pool database.Pool }

func NewLogisticsRepo(pool database.Pool) *LogisticsRepo { return &LogisticsRepo{pool: pool} }

func (r *LogisticsRepo) List(ctx context.Context, projectID int32) ([]models.Logistics, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "projectId", "materialName", unit, volume::text, "unitPrice"::text, "totalCost"::text,
		date, description, "createdAt", "updatedAt"
		FROM logistics WHERE "projectId" = $1 ORDER BY id DESC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Logistics
	for rows.Next() {
		var l models.Logistics
		var vol, hs, tb string
		var date sql.NullTime
		var descStr sql.NullString
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.MaterialName, &l.Unit, &vol, &hs, &tb, &date, &descStr, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		l.Volume = scanDec(vol)
		l.UnitPrice = scanDec(hs)
		l.TotalCost = scanDec(tb)
		if date.Valid {
			t := date.Time
			l.Date = &t
		}
		l.Description = strPtr(descStr)
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *LogisticsRepo) Create(ctx context.Context, projectID int32, materialName, unit string, volume, unitPrice decimal.Decimal, date, description *string, recordExpense bool) (*models.Logistics, *models.Transaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	totalCost := volume.Mul(unitPrice)
	var l models.Logistics
	var vol, hs, tb string
	var tg sql.NullTime
	var descStr sql.NullString
	if err := tx.QueryRow(ctx, `
		INSERT INTO logistics ("projectId", "materialName", unit, volume, "unitPrice", "totalCost", date, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, "projectId", "materialName", unit, volume::text, "unitPrice"::text, "totalCost"::text, date, description, "createdAt", "updatedAt"`,
		projectID, materialName, unit, decArg(volume), decArg(unitPrice), decArg(totalCost), date, description).Scan(&l.ID, &l.ProjectID, &l.MaterialName, &l.Unit, &vol, &hs, &tb, &tg, &descStr, &l.CreatedAt, &l.UpdatedAt); err != nil {
		return nil, nil, err
	}
	l.Volume = scanDec(vol)
	l.UnitPrice = scanDec(hs)
	l.TotalCost = scanDec(tb)
	if tg.Valid {
		t := tg.Time
		l.Date = &t
	}
	l.Description = strPtr(descStr)

	var re *models.Transaction
	if recordExpense {
		re, err = scanTransaction(tx.QueryRow(ctx, `
			INSERT INTO transactions ("projectId", date, category, amount, description, type, status, "logisticsId")
			VALUES ($1,$2,'Material',$3,$4,'expense','draft',$5)
			RETURNING id, "projectId", date, category, amount::text, description, type, status, "logisticsId", "invoiceId", "createdAt", "updatedAt"`,
			projectID, date, decArg(totalCost), description, l.ID))
		if err != nil {
			return nil, nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &l, re, nil
}

// --- Transaction ---
type TransactionRepo struct{ pool database.Pool }

func NewTransactionRepo(pool database.Pool) *TransactionRepo { return &TransactionRepo{pool: pool} }

func (r *TransactionRepo) List(ctx context.Context, projectID int32) ([]models.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "projectId", date, category, amount::text, description, type, status, "logisticsId", "invoiceId", "createdAt", "updatedAt"
		FROM transactions WHERE "projectId" = $1 ORDER BY date DESC, id DESC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Transaction
	for rows.Next() {
		var re models.Transaction
		var amountStr string
		var descStr sql.NullString
		if err := rows.Scan(&re.ID, &re.ProjectID, &re.Date, &re.Category, &amountStr, &descStr, &re.Type, &re.Status, &re.LogisticsID, &re.InvoiceID, &re.CreatedAt, &re.UpdatedAt); err != nil {
			return nil, err
		}
		re.Amount = scanDec(amountStr)
		re.Description = strPtr(descStr)
		out = append(out, re)
	}
	return out, rows.Err()
}

func (r *TransactionRepo) Create(ctx context.Context, projectID int32, date, category string, amount decimal.Decimal, description *string, transType models.TransactionType, status models.TransactionStatus, logisticsID, invoiceID *int32) (*models.Transaction, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO transactions ("projectId", date, category, amount, description, type, status, "logisticsId", "invoiceId")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, "projectId", date, category, amount::text, description, type, status, "logisticsId", "invoiceId", "createdAt", "updatedAt"`,
		projectID, date, category, decArg(amount), description, transType, status, logisticsID, invoiceID)
	var re models.Transaction
	var amountStr string
	var descStr sql.NullString
	if err := row.Scan(&re.ID, &re.ProjectID, &re.Date, &re.Category, &amountStr, &descStr, &re.Type, &re.Status, &re.LogisticsID, &re.InvoiceID, &re.CreatedAt, &re.UpdatedAt); err != nil {
		return nil, err
	}
	re.Amount = scanDec(amountStr)
	re.Description = strPtr(descStr)
	return &re, nil
}

const transactionSelect = `SELECT id, "projectId", date, category, amount::text, description, type, status, "logisticsId", "invoiceId", "createdAt", "updatedAt" FROM transactions`

func scanTransaction(row pgx.Row) (*models.Transaction, error) {
	var re models.Transaction
	var amountStr string
	var descStr sql.NullString
	if err := row.Scan(&re.ID, &re.ProjectID, &re.Date, &re.Category, &amountStr, &descStr, &re.Type, &re.Status, &re.LogisticsID, &re.InvoiceID, &re.CreatedAt, &re.UpdatedAt); err != nil {
		return nil, err
	}
	re.Amount = scanDec(amountStr)
	re.Description = strPtr(descStr)
	return &re, nil
}

func (r *TransactionRepo) Approve(ctx context.Context, projectID, id int32) (*models.Transaction, error) {
	return scanTransaction(r.pool.QueryRow(ctx, `
		UPDATE transactions SET status = 'approved', "updatedAt" = NOW()
		WHERE id = $1 AND "projectId" = $2 AND status = 'draft'
		RETURNING id, "projectId", date, category, amount::text, description, type, status, "logisticsId", "invoiceId", "createdAt", "updatedAt"`, id, projectID))
}

func (r *TransactionRepo) DeleteDraft(ctx context.Context, projectID, id int32) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM transactions WHERE id = $1 AND "projectId" = $2 AND status = 'draft'`, id, projectID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *TransactionRepo) FindByInvoiceID(ctx context.Context, projectID, invoiceID int32) (*models.Transaction, error) {
	return scanTransaction(r.pool.QueryRow(ctx, transactionSelect+` WHERE "invoiceId" = $1 AND "projectId" = $2 LIMIT 1`, invoiceID, projectID))
}

func (r *TransactionRepo) TagReverted(ctx context.Context, projectID, id int32) error {
	_, err := r.pool.Exec(ctx, `UPDATE transactions SET status = 'reverted', "updatedAt" = NOW() WHERE id = $1 AND "projectId" = $2`, id, projectID)
	return err
}

// --- Feedback ---

type FeedbackRepo struct{ pool database.Pool }

func NewFeedbackRepo(pool database.Pool) *FeedbackRepo { return &FeedbackRepo{pool: pool} }

func (r *FeedbackRepo) ListByUser(ctx context.Context, userID int32) ([]models.Feedback, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, "userId", subject, message, status, "createdAt", "updatedAt" FROM feedback WHERE "userId" = $1 ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Feedback
	for rows.Next() {
		var f models.Feedback
		if err := rows.Scan(&f.ID, &f.UserID, &f.Subject, &f.Message, &f.Status, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FeedbackRepo) Create(ctx context.Context, userID int32, subject, message string) (*models.Feedback, error) {
	var f models.Feedback
	err := r.pool.QueryRow(ctx, `
		INSERT INTO feedback ("userId", subject, message) VALUES ($1,$2,$3)
		RETURNING id, "userId", subject, message, status, "createdAt", "updatedAt"`, userID, subject, message).
		Scan(&f.ID, &f.UserID, &f.Subject, &f.Message, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}
