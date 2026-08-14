package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

// --- RecapRepo (margin) ---

func TestRecapGetMargin(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`SELECT margin::text FROM recaps`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"margin"}).AddRow("0.15"))
	r := NewRecapRepo(m)
	g, err := r.GetMargin(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMargin: %v", err)
	}
	if !g.Equal(decimal.RequireFromString("0.15")) {
		t.Errorf("margin = %s", g)
	}
}

func TestRecapGetMarginNone(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`SELECT margin::text FROM recaps`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"margin"}).AddRow(nil))
	r := NewRecapRepo(m)
	g, err := r.GetMargin(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMargin: %v", err)
	}
	if !g.IsZero() {
		t.Errorf("margin = %s want zero", g)
	}
}

func TestRecapUpsertMargin(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`INSERT INTO recaps`).
		WithArgs(int32(1), "0.12").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	r := NewRecapRepo(m)
	if err := r.UpsertMargin(context.Background(), 1, decimal.RequireFromString("0.12")); err != nil {
		t.Fatalf("UpsertMargin: %v", err)
	}
}

// --- InvoiceRepo ---

func invoiceRow() []string {
	return []string{"id", "projectId", "number", "date", "dueDate", "poNumber", "buyerName", "buyerAddress", "buyerContact",
		"discount", "taxRate", "paymentBank", "paymentAccountNumber", "paymentAccountName", "notes", "financeName",
		"total", "status", "createdAt", "updatedAt"}
}

func TestInvoiceListWithItems(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM invoices WHERE "projectId"`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(invoiceRow()).
			AddRow(int32(1), int32(1), "INV/2026/01/001", time.Now(), nil, nil, "PT Pembeli", "Jl. A", nil,
				"0", "0.11", nil, nil, nil, nil, "Andi", "1100000", models.InvoiceDraft, time.Now(), time.Now()))
	m.ExpectQuery(`FROM invoice_items WHERE`).WithArgs([]int32{1}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "invoiceId", "description", "qty", "unit", "unitPrice"}).
			AddRow(int32(1), int32(1), "Galian", "10", "m3", "100000"))
	r := NewInvoiceRepo(m)
	invs, err := r.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(invs) != 1 {
		t.Fatalf("invs = %d", len(invs))
	}
	if invs[0].Number != "INV/2026/01/001" {
		t.Errorf("number = %q", invs[0].Number)
	}
	if !invs[0].Total.Equal(decimal.NewFromInt(1100000)) {
		t.Errorf("total = %s", invs[0].Total)
	}
	if len(invs[0].Items) != 1 || !invs[0].Items[0].Qty.Equal(decimal.NewFromInt(10)) {
		t.Errorf("items = %+v", invs[0].Items)
	}
}

func TestInvoiceGet(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM invoices WHERE id =`).WithArgs(int32(1), int32(1)).
		WillReturnRows(pgxmock.NewRows(invoiceRow()).
			AddRow(int32(1), int32(1), "INV/2026/01/001", time.Now(), nil, nil, "PT Pembeli", "Jl. A", nil,
				"0", "0.11", nil, nil, nil, nil, "Andi", "1100000", models.InvoiceDraft, time.Now(), time.Now()))
	m.ExpectQuery(`FROM invoice_items WHERE`).WithArgs([]int32{1}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "invoiceId", "description", "qty", "unit", "unitPrice"}))
	r := NewInvoiceRepo(m)
	inv, err := r.Get(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if inv.Status != models.InvoiceDraft {
		t.Errorf("status = %s", inv.Status)
	}
}

func TestInvoiceCreate(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`SELECT COALESCE\(MAX`).WithArgs(int32(1), "2026-08-11").
		WillReturnRows(pgxmock.NewRows([]string{"seq"}).AddRow(2))
	m.ExpectBegin()
	m.ExpectQuery(`INSERT INTO invoices`).
		WithArgs(int32(1), "INV/2026/08/002", "2026-08-11", (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil),
			"0", "0", (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil), "5000000", models.InvoiceDraft).
		WillReturnRows(pgxmock.NewRows(invoiceRow()).
			AddRow(int32(3), int32(1), "INV/2026/08/002", time.Now(), nil, nil, nil, nil, nil,
				"0", "0", nil, nil, nil, nil, nil, "5000000", models.InvoiceDraft, time.Now(), time.Now()))
	m.ExpectExec(`INSERT INTO invoice_items`).
		WithArgs(int32(3), "Pekerjaan A", "1", "ls", "5000000").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectCommit()

	r := NewInvoiceRepo(m)
	inv, err := r.Create(context.Background(), 1, &InvoiceInput{
		Date:   "2026-08-11",
		Total:  decimal.NewFromInt(5000000),
		Status: models.InvoiceDraft,
		Items:  []models.InvoiceItem{{Description: "Pekerjaan A", Qty: decimal.NewFromInt(1), Unit: "ls", UnitPrice: decimal.NewFromInt(5000000)}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if inv.ID != 3 || inv.Number != "INV/2026/08/002" {
		t.Errorf("inv = %+v", inv)
	}
	if len(inv.Items) != 1 {
		t.Errorf("items = %+v", inv.Items)
	}
}

func TestInvoiceUpdateStatus(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`UPDATE invoices SET status`).WithArgs(int32(1), int32(1), models.InvoicePaid).
		WillReturnRows(pgxmock.NewRows(invoiceRow()).
			AddRow(int32(1), int32(1), "INV/2026/01/001", time.Now(), nil, nil, nil, nil, nil,
				"0", "0", nil, nil, nil, nil, nil, "1100000", models.InvoicePaid, time.Now(), time.Now()))
	r := NewInvoiceRepo(m)
	inv, err := r.UpdateStatus(context.Background(), 1, 1, models.InvoicePaid)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if inv.Status != models.InvoicePaid {
		t.Errorf("status = %s", inv.Status)
	}
}

// --- LogisticsRepo ---

func TestLogisticsList(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM logistics WHERE`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "materialName", "unit", "volume", "unitPrice", "totalCost",
			"date", "description", "createdAt", "updatedAt"}).
			AddRow(int32(1), int32(1), "Semen", "kg", "500", "1750", "875000", time.Now(), "pesanan 1", time.Now(), time.Now()))
	r := NewLogisticsRepo(m)
	ls, err := r.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ls) != 1 || !ls[0].TotalCost.Equal(decimal.NewFromInt(875000)) {
		t.Fatalf("ls = %+v", ls)
	}
	if ls[0].Description == nil || *ls[0].Description != "pesanan 1" {
		t.Errorf("desc = %v", ls[0].Description)
	}
}

func TestLogisticsCreateWithExpense(t *testing.T) {
	m := newPool(t)
	m.ExpectBegin()
	m.ExpectQuery(`INSERT INTO logistics`).
		WithArgs(int32(1), "Semen", "kg", "500", "1750", "875000", (*string)(nil), (*string)(nil)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "materialName", "unit", "volume", "unitPrice", "totalCost",
			"date", "description", "createdAt", "updatedAt"}).
			AddRow(int32(1), int32(1), "Semen", "kg", "500", "1750", "875000", nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`INSERT INTO transactions`).
		WithArgs(int32(1), (*string)(nil), "875000", (*string)(nil), int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "date", "category", "amount", "description", "type", "status",
			"logisticsId", "invoiceId", "createdAt", "updatedAt"}).
			AddRow(int32(5), int32(1), time.Now(), "Material", "875000", nil, models.TransactionExpense, models.TransactionDraft, int32Ptr(1), nil, time.Now(), time.Now()))
	m.ExpectCommit()

	r := NewLogisticsRepo(m)
	l, re, err := r.Create(context.Background(), 1, "Semen", "kg",
		decimal.NewFromInt(500), decimal.NewFromInt(1750), nil, nil, true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !l.TotalCost.Equal(decimal.NewFromInt(875000)) {
		t.Errorf("totalCost = %s", l.TotalCost)
	}
	if re == nil || re.ID != 5 || re.Type != models.TransactionExpense {
		t.Errorf("txn = %+v", re)
	}
}

func TestLogisticsCreateWithoutExpense(t *testing.T) {
	m := newPool(t)
	m.ExpectBegin()
	m.ExpectQuery(`INSERT INTO logistics`).
		WithArgs(int32(1), "Semen", "kg", "500", "1750", "875000", (*string)(nil), (*string)(nil)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "materialName", "unit", "volume", "unitPrice", "totalCost",
			"date", "description", "createdAt", "updatedAt"}).
			AddRow(int32(1), int32(1), "Semen", "kg", "500", "1750", "875000", nil, nil, time.Now(), time.Now()))
	m.ExpectCommit()
	r := NewLogisticsRepo(m)
	_, re, err := r.Create(context.Background(), 1, "Semen", "kg",
		decimal.NewFromInt(500), decimal.NewFromInt(1750), nil, nil, false)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if re != nil {
		t.Errorf("re = %+v want nil", re)
	}
}

// --- TransactionRepo ---

func transactionRow() []string {
	return []string{"id", "projectId", "date", "category", "amount", "description", "type", "status",
		"logisticsId", "invoiceId", "createdAt", "updatedAt"}
}

func TestTransactionList(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM transactions WHERE "projectId"`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(transactionRow()).
			AddRow(int32(1), int32(1), time.Now(), "Material", "500000", nil, models.TransactionExpense, models.TransactionApproved, nil, nil, time.Now(), time.Now()))
	r := NewTransactionRepo(m)
	ts, err := r.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ts) != 1 || !ts[0].Amount.Equal(decimal.NewFromInt(500000)) {
		t.Fatalf("ts = %+v", ts)
	}
	if ts[0].Status != models.TransactionApproved {
		t.Errorf("status = %s", ts[0].Status)
	}
}

func TestTransactionCreate(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`INSERT INTO transactions`).
		WithArgs(int32(1), "2026-08-11", "Material", "750000", (*string)(nil), models.TransactionExpense, models.TransactionDraft, (*int32)(nil), (*int32)(nil)).
		WillReturnRows(pgxmock.NewRows(transactionRow()).
			AddRow(int32(1), int32(1), time.Now(), "Material", "750000", nil, models.TransactionExpense, models.TransactionDraft, nil, nil, time.Now(), time.Now()))
	r := NewTransactionRepo(m)
	re, err := r.Create(context.Background(), 1, "2026-08-11", "Material", decimal.NewFromInt(750000), nil,
		models.TransactionExpense, models.TransactionDraft, nil, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if re.ID != 1 {
		t.Errorf("id = %d", re.ID)
	}
}

func TestTransactionApprove(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`UPDATE transactions SET status`).WithArgs(int32(1), int32(1)).
		WillReturnRows(pgxmock.NewRows(transactionRow()).
			AddRow(int32(1), int32(1), time.Now(), "Material", "750000", nil, models.TransactionExpense, models.TransactionApproved, nil, nil, time.Now(), time.Now()))
	r := NewTransactionRepo(m)
	re, err := r.Approve(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Approve: %v", err)
	}
	if re.Status != models.TransactionApproved {
		t.Errorf("status = %s", re.Status)
	}
}

func TestTransactionDeleteDraft(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`DELETE FROM transactions`).WithArgs(int32(1), int32(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	r := NewTransactionRepo(m)
	ok, err := r.DeleteDraft(context.Background(), 1, 1)
	if err != nil || !ok {
		t.Fatalf("DeleteDraft = %v err %v", ok, err)
	}
}

func TestTransactionDeleteDraftNotDraft(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`DELETE FROM transactions`).WithArgs(int32(1), int32(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	r := NewTransactionRepo(m)
	ok, err := r.DeleteDraft(context.Background(), 1, 1)
	if err != nil || ok {
		t.Fatalf("DeleteDraft = %v err %v", ok, err)
	}
}

func TestTransactionFindByInvoiceID(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM transactions WHERE "invoiceId"`).WithArgs(int32(3), int32(1)).
		WillReturnRows(pgxmock.NewRows(transactionRow()).
			AddRow(int32(1), int32(1), time.Now(), "Invoice", "1100000", nil, models.TransactionIncome, models.TransactionDraft, nil, int32Ptr(3), time.Now(), time.Now()))
	r := NewTransactionRepo(m)
	re, err := r.FindByInvoiceID(context.Background(), 1, 3)
	if err != nil {
		t.Fatalf("FindByInvoiceID: %v", err)
	}
	if re.InvoiceID == nil || *re.InvoiceID != 3 {
		t.Errorf("invoiceId = %v", re.InvoiceID)
	}
}

func TestTransactionTagReverted(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`UPDATE transactions SET status = 'reverted'`).WithArgs(int32(1), int32(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	r := NewTransactionRepo(m)
	if err := r.TagReverted(context.Background(), 1, 1); err != nil {
		t.Fatalf("TagReverted: %v", err)
	}
}

// --- FeedbackRepo ---

func TestFeedbackListByUser(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM feedback WHERE "userId"`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "userId", "subject", "message", "status", "createdAt", "updatedAt"}).
			AddRow(int32(1), int32(1), "Bug", "ada bug", models.FeedbackOpen, time.Now(), time.Now()))
	r := NewFeedbackRepo(m)
	fs, err := r.ListByUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(fs) != 1 || fs[0].Status != models.FeedbackOpen {
		t.Fatalf("fs = %+v", fs)
	}
}

func TestFeedbackCreate(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`INSERT INTO feedback`).WithArgs(int32(1), "Saran", "tambah fitur").
		WillReturnRows(pgxmock.NewRows([]string{"id", "userId", "subject", "message", "status", "createdAt", "updatedAt"}).
			AddRow(int32(1), int32(1), "Saran", "tambah fitur", models.FeedbackOpen, time.Now(), time.Now()))
	r := NewFeedbackRepo(m)
	f, err := r.Create(context.Background(), 1, "Saran", "tambah fitur")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if f.Subject != "Saran" {
		t.Errorf("subject = %q", f.Subject)
	}
}
