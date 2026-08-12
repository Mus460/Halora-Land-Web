package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
	"github.com/halora-land/halora-be/service"
)

type ProjectSubHandler struct {
	pool     database.Pool
	projects *repository.ProjectRepo
	recaps   *repository.RecapRepo
	rab      *service.RABService
	snap     *service.SnapshotService
	real     *repository.TransactionRepo
	log      *repository.LogisticsRepo
	inv      *repository.InvoiceRepo
}

func NewProjectSubHandler(pool database.Pool, pr *repository.ProjectRepo, rr *repository.RecapRepo, rab *service.RABService, ss *service.SnapshotService) *ProjectSubHandler {
	return &ProjectSubHandler{
		pool: pool, projects: pr, recaps: rr, rab: rab, snap: ss,
		real: repository.NewTransactionRepo(pool),
		log:  repository.NewLogisticsRepo(pool),
		inv:  repository.NewInvoiceRepo(pool),
	}
}

func (h *ProjectSubHandler) RecapGet(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	summary, err := h.projects.Summary(r.Context(), pid)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	res, err := h.rab.Compute(r.Context(), pid, summary)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *ProjectSubHandler) RecapPut(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		Margin decimal.Decimal `json:"margin"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := h.recaps.UpsertMargin(r.Context(), pid, in.Margin); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "margin": in.Margin})
}

func (h *ProjectSubHandler) RecalculateAll(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	u := auth.FromContext(r.Context())
	n, err := h.snap.RecalculateAll(r.Context(), pid, u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "recalculated": n})
}

func (h *ProjectSubHandler) TransactionList(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	out, err := h.real.List(r.Context(), pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": out})
}

func (h *ProjectSubHandler) TransactionCreate(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		Date        string                 `json:"date"`
		Category    string                 `json:"category"`
		Amount      decimal.Decimal        `json:"amount"`
		Description *string                `json:"description"`
		Type        models.TransactionType `json:"type"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Date == "" {
		writeError(w, http.StatusBadRequest, "date wajib diisi")
		return
	}
	if in.Category == "" {
		writeError(w, http.StatusBadRequest, "category wajib diisi")
		return
	}
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		writeError(w, http.StatusBadRequest, "amount harus lebih dari 0")
		return
	}
	if in.Type == "" {
		in.Type = models.TransactionExpense
	}
	if in.Type != models.TransactionExpense && in.Type != models.TransactionIncome {
		writeError(w, http.StatusBadRequest, "type tidak valid")
		return
	}
	out, err := h.real.Create(r.Context(), pid, in.Date, in.Category, in.Amount, in.Description, in.Type, models.TransactionApproved, nil, nil)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"transactions": out})
}

func (h *ProjectSubHandler) TransactionApprove(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	rid, ok := parseIntParam(w, r, "transactionId")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		Status models.TransactionStatus `json:"status"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Status != models.TransactionApproved {
		writeError(w, http.StatusBadRequest, "status tidak valid")
		return
	}
	out, err := h.real.Approve(r.Context(), pid, rid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Transaksi tidak ditemukan atau sudah disetujui")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": out})
}

func (h *ProjectSubHandler) TransactionDelete(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	rid, ok := parseIntParam(w, r, "transactionId")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	deleted, err := h.real.DeleteDraft(r.Context(), pid, rid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "Transaksi tidak ditemukan atau tidak bisa dihapus")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *ProjectSubHandler) LogisticsList(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	out, err := h.log.List(r.Context(), pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"logistics": out})
}

func (h *ProjectSubHandler) LogisticsCreate(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		MaterialName  string          `json:"materialName"`
		Unit          string          `json:"unit"`
		Volume        decimal.Decimal `json:"volume"`
		UnitPrice     decimal.Decimal `json:"unitPrice"`
		Date          *string         `json:"date"`
		Description   *string         `json:"description"`
		RecordExpense bool            `json:"recordExpense"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.MaterialName == "" {
		writeError(w, http.StatusBadRequest, "name material wajib diisi")
		return
	}
	if in.Unit == "" {
		writeError(w, http.StatusBadRequest, "unit wajib diisi")
		return
	}
	if in.Volume.LessThanOrEqual(decimal.Zero) {
		writeError(w, http.StatusBadRequest, "volume harus lebih dari 0")
		return
	}
	if in.UnitPrice.IsNegative() {
		writeError(w, http.StatusBadRequest, "price unit tidak boleh negatif")
		return
	}
	if in.RecordExpense && in.Date == nil {
		today := time.Now().Format("2006-01-02")
		in.Date = &today
	}
	out, linked, err := h.log.Create(r.Context(), pid, in.MaterialName, in.Unit, in.Volume, in.UnitPrice, in.Date, in.Description, in.RecordExpense)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"logistics": out, "transactions": linked})
}

func (h *ProjectSubHandler) InvoiceList(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	out, err := h.inv.List(r.Context(), pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"invoices": out})
}

func (h *ProjectSubHandler) InvoiceCreate(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		Date                 string               `json:"date"`
		DueDate              *string              `json:"dueDate"`
		PONumber             *string              `json:"poNumber"`
		BuyerName            *string              `json:"buyerName"`
		BuyerAddress         *string              `json:"buyerAddress"`
		BuyerContact         *string              `json:"buyerContact"`
		Discount             decimal.Decimal      `json:"discount"`
		TaxRate              decimal.Decimal      `json:"taxRate"`
		PaymentBank          *string              `json:"paymentBank"`
		PaymentAccountNumber *string              `json:"paymentAccountNumber"`
		PaymentAccountName   *string              `json:"paymentAccountName"`
		Notes                *string              `json:"notes"`
		FinanceName          *string              `json:"financeName"`
		Total                decimal.Decimal      `json:"total"`
		Status               models.InvoiceStatus `json:"status"`
		Items                []models.InvoiceItem `json:"items"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Date == "" {
		writeError(w, http.StatusBadRequest, "date wajib diisi")
		return
	}
	if in.Status == "" {
		in.Status = models.InvoiceDraft
	}
	switch in.Status {
	case models.InvoiceDraft, models.InvoiceSent, models.InvoicePaid:
	default:
		writeError(w, http.StatusBadRequest, "status tidak valid")
		return
	}
	if len(in.Items) == 0 {
		if in.Total.LessThanOrEqual(decimal.Zero) {
			writeError(w, http.StatusBadRequest, "total harus lebih dari 0")
			return
		}
	} else {
		if err := validateInvoiceItems(in.Items); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		in.Total = invoiceGrandTotal(in.Items, in.Discount, in.TaxRate)
	}
	out, err := h.inv.Create(r.Context(), pid, &repository.InvoiceInput{
		Date:                 in.Date,
		DueDate:              in.DueDate,
		PONumber:             in.PONumber,
		BuyerName:            in.BuyerName,
		BuyerAddress:         in.BuyerAddress,
		BuyerContact:         in.BuyerContact,
		Discount:             in.Discount,
		TaxRate:              in.TaxRate,
		PaymentBank:          in.PaymentBank,
		PaymentAccountNumber: in.PaymentAccountNumber,
		PaymentAccountName:   in.PaymentAccountName,
		Notes:                in.Notes,
		FinanceName:          in.FinanceName,
		Total:                in.Total,
		Status:               in.Status,
		Items:                in.Items,
	})
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"invoices": out})
}

func validateInvoiceItems(items []models.InvoiceItem) error {
	for _, it := range items {
		if it.Description == "" {
			return errors.New("deskripsi item wajib diisi")
		}
		if it.Qty.LessThanOrEqual(decimal.Zero) {
			return errors.New("qty harus lebih dari 0")
		}
		if it.Unit == "" {
			return errors.New("unit wajib diisi")
		}
		if it.UnitPrice.IsNegative() {
			return errors.New("harga satuan tidak boleh negatif")
		}
	}
	return nil
}

func invoiceGrandTotal(items []models.InvoiceItem, discount, taxRate decimal.Decimal) decimal.Decimal {
	subtotal := decimal.Zero
	for _, it := range items {
		subtotal = subtotal.Add(it.Qty.Mul(it.UnitPrice))
	}
	base := subtotal.Sub(discount)
	if base.IsNegative() {
		base = decimal.Zero
	}
	tax := base.Mul(taxRate).Div(decimal.NewFromInt(100))
	return base.Add(tax).Round(2)
}

func (h *ProjectSubHandler) InvoiceUpdate(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	iid, ok := parseIntParam(w, r, "invoiceId")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		Date                 *string               `json:"date"`
		DueDate              *string               `json:"dueDate"`
		PONumber             *string               `json:"poNumber"`
		BuyerName            *string               `json:"buyerName"`
		BuyerAddress         *string               `json:"buyerAddress"`
		BuyerContact         *string               `json:"buyerContact"`
		Discount             *decimal.Decimal      `json:"discount"`
		TaxRate              *decimal.Decimal      `json:"taxRate"`
		PaymentBank          *string               `json:"paymentBank"`
		PaymentAccountNumber *string               `json:"paymentAccountNumber"`
		PaymentAccountName   *string               `json:"paymentAccountName"`
		Notes                *string               `json:"notes"`
		FinanceName          *string               `json:"financeName"`
		Status               *models.InvoiceStatus `json:"status"`
		RecordExpense        bool                  `json:"recordExpense"`
		Items                []models.InvoiceItem  `json:"items"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	current, err := h.inv.Get(r.Context(), pid, iid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Invoice tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hasDocumentFields := in.Date != nil || in.DueDate != nil || in.PONumber != nil ||
		in.BuyerName != nil || in.BuyerAddress != nil || in.BuyerContact != nil ||
		in.Discount != nil || in.TaxRate != nil || in.PaymentBank != nil ||
		in.PaymentAccountNumber != nil || in.PaymentAccountName != nil || in.Notes != nil ||
		in.FinanceName != nil || in.Items != nil
	if !hasDocumentFields && in.Status == nil {
		writeError(w, http.StatusBadRequest, "tidak ada data yang diubah")
		return
	}

	var inv *models.Invoice
	if hasDocumentFields {
		// Document fields are only editable while the invoice is a draft.
		if current.Status != models.InvoiceDraft {
			writeError(w, http.StatusBadRequest, "Invoice sudah terkirim/lunas, tidak bisa diedit")
			return
		}
		date := current.Date.Format("2006-01-02")
		if in.Date != nil {
			if *in.Date == "" {
				writeError(w, http.StatusBadRequest, "date wajib diisi")
				return
			}
			date = *in.Date
		}
		discount := current.Discount
		if in.Discount != nil {
			discount = *in.Discount
		}
		taxRate := current.TaxRate
		if in.TaxRate != nil {
			taxRate = *in.TaxRate
		}
		items := current.Items
		if in.Items != nil {
			if err := validateInvoiceItems(in.Items); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			items = in.Items
		}
		total := invoiceGrandTotal(items, discount, taxRate)
		inv, err = h.inv.Update(r.Context(), pid, iid, &repository.InvoiceInput{
			Date:                 date,
			DueDate:              firstNonNil(in.DueDate, ptrFromTime(current.DueDate)),
			PONumber:             firstNonNil(in.PONumber, current.PONumber),
			BuyerName:            firstNonNil(in.BuyerName, current.BuyerName),
			BuyerAddress:         firstNonNil(in.BuyerAddress, current.BuyerAddress),
			BuyerContact:         firstNonNil(in.BuyerContact, current.BuyerContact),
			Discount:             discount,
			TaxRate:              taxRate,
			PaymentBank:          firstNonNil(in.PaymentBank, current.PaymentBank),
			PaymentAccountNumber: firstNonNil(in.PaymentAccountNumber, current.PaymentAccountNumber),
			PaymentAccountName:   firstNonNil(in.PaymentAccountName, current.PaymentAccountName),
			Notes:                firstNonNil(in.Notes, current.Notes),
			FinanceName:          firstNonNil(in.FinanceName, current.FinanceName),
			Total:                total,
			Items:                items,
		})
	} else {
		status := current.Status
		if in.Status != nil {
			switch *in.Status {
			case models.InvoiceDraft, models.InvoiceSent, models.InvoicePaid:
				status = *in.Status
			default:
				writeError(w, http.StatusBadRequest, "status tidak valid")
				return
			}
		}
		inv, err = h.inv.UpdateStatus(r.Context(), pid, iid, status)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Invoice tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var linked *models.Transaction
	if inv.Status == models.InvoicePaid && current.Status != models.InvoicePaid {
		existing, err := h.real.FindByInvoiceID(r.Context(), pid, iid)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			date := inv.Date.Format("2006-01-02")
			ket := inv.Number
			linked, err = h.real.Create(r.Context(), pid, date, "Invoice", inv.Total, &ket, models.TransactionIncome, models.TransactionDraft, nil, &iid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			linked = existing
		}
	} else if inv.Status != models.InvoicePaid {
		existing, err := h.real.FindByInvoiceID(r.Context(), pid, iid)
		if err == nil && existing.Status != models.TransactionReverted {
			if err := h.real.TagReverted(r.Context(), pid, existing.ID); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			existing.Status = models.TransactionReverted
			linked = existing
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"invoices": inv, "transactions": linked})
}

func firstNonNil[T any](a, b *T) *T {
	if a != nil {
		return a
	}
	return b
}

func ptrFromTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func (h *ProjectSubHandler) SCurve(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	items, err := h.rab.Compute(r.Context(), pid, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(items.Grouped) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{
			"labels":  []string{},
			"planned": []int{},
			"actual":  []int{},
		})
		return
	}
	labels, planned, actual := h.sCurveSeries(r.Context(), pid)
	writeJSON(w, http.StatusOK, map[string]any{
		"labels":  labels,
		"planned": planned,
		"actual":  actual,
	})
}

var sCurveMonths = []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// sCurveSeries builds the cumulative S-curve from work item progress logs:
// each log entry is a timestamped state change, and its cost-weighted delta
// (progress - previous progress, % of totalCost) lands in the month the log was
// recorded. The planned line ramps 0-100% across the project's timeline
// (timelineMonths + timelineDays); labels run from the project start month
// through the longer of the elapsed or planned span.
func (h *ProjectSubHandler) sCurveSeries(ctx context.Context, pid int32) ([]string, []int, []int) {
	var start time.Time
	var tMonths, tDays int
	if err := h.pool.QueryRow(ctx, `SELECT "createdAt", "timelineMonths", "timelineDays" FROM projects WHERE id = $1`, pid).Scan(&start, &tMonths, &tDays); err != nil {
		return []string{}, []int{}, []int{}
	}
	now := time.Now()
	elapsed := monthsBetween(start, now) + 1
	if elapsed < 1 {
		elapsed = 1
	}
	duration := float64(tMonths) + float64(tDays)/30.0
	span := elapsed
	if duration > 0 {
		if want := int(math.Ceil(duration)); want > span {
			span = want
		}
	}
	labels := make([]string, span)
	for i := 0; i < span; i++ {
		m := start.AddDate(0, i, 0)
		labels[i] = fmt.Sprintf("%s %02d", sCurveMonths[int(m.Month())-1], m.Year()%100)
	}
	rows, err := h.pool.Query(ctx, `
		SELECT id, "totalCost"::text, "createdAt"
		FROM work_items WHERE "projectId" = $1 AND "deletedAt" IS NULL`, pid)
	if err != nil {
		return labels, linearPlanned(span), make([]int, span)
	}
	defer rows.Close()
	var monthly [12]decimal.Decimal
	var totalCost decimal.Decimal
	for rows.Next() {
		var itemID int32
		var costStr string
		var itemCreated time.Time
		if err := rows.Scan(&itemID, &costStr, &itemCreated); err != nil {
			continue
		}
		cost := decimal.RequireFromString(costStr)
		totalCost = totalCost.Add(cost)
		logs, err := h.pool.Query(ctx, `
			SELECT progress, "createdAt" FROM work_item_progress_logs
			WHERE "workItemId" = $1 ORDER BY "createdAt" ASC, id ASC`, itemID)
		if err != nil {
			continue
		}
		type logEntry struct {
			progress int
			at       time.Time
		}
		var entries []logEntry
		for logs.Next() {
			var pr int
			var at time.Time
			if err := logs.Scan(&pr, &at); err != nil {
				break
			}
			entries = append(entries, logEntry{progress: pr, at: at})
		}
		logs.Close()
		if len(entries) == 0 {
			if cost.IsZero() {
				continue
			}
			entries = []logEntry{{progress: 0, at: itemCreated}}
		}
		prev := 0
		for _, e := range entries {
			if e.progress <= prev {
				prev = e.progress
				continue
			}
			delta := cost.Mul(decimal.NewFromInt(int64(e.progress - prev))).Div(decimal.NewFromInt(100))
			idx := monthsBetween(start, e.at)
			if idx < 0 {
				idx = 0
			}
			if idx >= span {
				idx = span - 1
			}
			monthly[idx] = monthly[idx].Add(delta)
			prev = e.progress
		}
	}
	actual := make([]int, span)
	if totalCost.IsPositive() {
		cum := decimal.Zero
		for i := 0; i < span; i++ {
			cum = cum.Add(monthly[i])
			pct := int(cum.Mul(decimal.NewFromInt(100)).Div(totalCost).IntPart())
			if pct > 100 {
				pct = 100
			}
			actual[i] = pct
		}
	}
	var planned []int
	if duration > 0 {
		planned = timelinePlanned(span, duration)
	} else {
		planned = linearPlanned(span)
	}
	return labels, planned, actual
}

// timelinePlanned ramps 0-100% linearly across a fractional month duration,
// so a 6-month-15-day timeline reaches 100% partway through the last bucket.
func timelinePlanned(span int, duration float64) []int {
	planned := make([]int, span)
	for i := range planned {
		pct := int(float64(i+1) / duration * 100)
		if pct > 100 {
			pct = 100
		}
		planned[i] = pct
	}
	if span > 0 {
		planned[span-1] = 100
	}
	return planned
}

func linearPlanned(span int) []int {
	planned := make([]int, span)
	for i := range planned {
		planned[i] = (i + 1) * (100 / span)
	}
	if span > 0 && planned[span-1] < 100 {
		planned[span-1] = 100
	}
	return planned
}

func monthsBetween(a, b time.Time) int {
	return (b.Year()-a.Year())*12 + int(b.Month()-a.Month())
}
