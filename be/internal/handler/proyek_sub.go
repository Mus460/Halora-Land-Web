package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
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
	progress *service.ProgressService
	real     *repository.TransactionRepo
	log      *repository.LogisticsRepo
	inv      *repository.InvoiceRepo
}

func NewProjectSubHandler(pool database.Pool, pr *repository.ProjectRepo, rr *repository.RecapRepo, rab *service.RABService, ss *service.SnapshotService, progress *service.ProgressService) *ProjectSubHandler {
	return &ProjectSubHandler{
		pool: pool, projects: pr, recaps: rr, rab: rab, snap: ss, progress: progress,
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

const (
	// hoursPerWeek converts work-item duration hours into schedule weeks
	// (5 workdays × 8h).
	hoursPerWeek = 40.0
	// sCurveWeekCap keeps pathological schedules from producing huge charts.
	sCurveWeekCap = 156
)

// sCurveCategoryRank sequences work items in the actual construction flow:
// persiapan first, then struktur, arsitektur, and finally MEP.
var sCurveCategoryRank = map[models.WorkCategory]int{
	models.CategoryPreparation: 1,
	models.CategoryFoundation:  2,
	models.CategoryConcrete:    3,
	models.CategoryCanopy:      4,
	models.CategorySteel:       5,
	models.CategoryStairs:      6,
	models.CategoryRoof:        7,
	models.CategoryWall:        8,
	models.CategoryPlastering:  9,
	models.CategoryFinishing:   10,
	models.CategoryTiles:       11,
	models.CategoryPaving:      12,
	models.CategoryPainting:    13,
	models.CategoryDoors:       14,
	models.CategoryInterior:    15,
	models.CategoryToilet:      16,
	models.CategoryMEP:         17,
	models.CategoryCustom:      18,
}

type curveItem struct {
	id     int32
	rank   int
	cost   decimal.Decimal
	weight float64
	hours  float64
	weeks  float64
	start  float64
}

// sCurveSeries builds the weekly S-curve series:
//
//   - Planned: work items are weighted by their cost share of the project
//     subtotal (w_i = totalCost_i / ΣtotalCost), sequenced in construction
//     flow order (persiapan → struktur → arsitektur → MEP), and each item
//     occupies its estimated duration in weeks (duration_hours × volume / 40h).
//     Items without an estimated duration receive a weight-proportional share
//     of the schedule slack (or a 0.5-week slot when durations fill the plan).
//     Cumulative percent is sampled at the end of every integer week.
//   - Actual: cost-weighted progress deltas from work item logs, bucketed per
//     week since project start.
//
// Both lines start at 0% at W0 (project start); labels run W0…WN where
// N = max(scheduled weeks, weeks elapsed), capped at sCurveWeekCap.
func (h *ProjectSubHandler) sCurveSeries(ctx context.Context, pid int32) ([]string, []int, []int) {
	var start time.Time
	var tMonths, tDays int
	if err := h.pool.QueryRow(ctx, `SELECT "createdAt", "timelineMonths", "timelineDays" FROM projects WHERE id = $1`, pid).Scan(&start, &tMonths, &tDays); err != nil {
		return []string{}, []int{}, []int{}
	}
	weighted, totalCost, err := h.progress.Items(ctx, pid)
	if err != nil {
		return []string{}, []int{}, []int{}
	}
	items := make([]curveItem, 0, len(weighted))
	for i := range weighted {
		w := &weighted[i]
		rank := sCurveCategoryRank[w.Category]
		if rank == 0 {
			rank = 100
		}
		items = append(items, curveItem{id: w.ID, rank: rank, cost: w.Cost, weight: w.Weight, hours: w.Hours})
	}
	if len(items) == 0 || !totalCost.IsPositive() {
		return []string{}, []int{}, []int{}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].rank != items[j].rank {
			return items[i].rank < items[j].rank
		}
		return items[i].id < items[j].id
	})

	// Schedule: items with an estimated duration occupy hours/40 weeks each;
	// the plan length is at least the project timeline. Items without a
	// duration split the slack proportional to their cost weight, or get a
	// 0.5-week slot when durations already fill the plan.
	durationWeeks := 0.0
	noDurWeight := decimal.Zero
	for i := range items {
		if items[i].hours > 0 {
			items[i].weeks = items[i].hours / hoursPerWeek
			durationWeeks += items[i].weeks
		} else {
			noDurWeight = noDurWeight.Add(items[i].cost)
		}
	}
	totalWeeks := durationWeeks
	if tl := float64(tMonths)*4.333 + float64(tDays)/7.0; tl > totalWeeks {
		totalWeeks = tl
	}
	if totalWeeks < 1 {
		totalWeeks = 1
	}
	slack := totalWeeks - durationWeeks
	if noDurWeight.IsPositive() && slack > 0.001 {
		for i := range items {
			if items[i].weeks == 0 {
				items[i].weeks, _ = items[i].cost.Div(noDurWeight).Mul(decimal.NewFromFloat(slack)).Float64()
			}
		}
	} else {
		for i := range items {
			if items[i].weeks == 0 {
				items[i].weeks = 0.5
			}
		}
	}

	cum := 0.0
	for i := range items {
		items[i].start = cum
		cum += items[i].weeks
	}
	elapsedWeeks := float64(daysBetween(start, time.Now())) / 7.0
	if elapsedWeeks < 0 {
		elapsedWeeks = 0
	}
	span := int(math.Ceil(cum))
	if int(math.Ceil(elapsedWeeks)) > span {
		span = int(math.Ceil(elapsedWeeks))
	}
	if span < 1 {
		span = 1
	}
	if span > sCurveWeekCap {
		span = sCurveWeekCap
	}

	labels := make([]string, span)
	for i := range labels {
		labels[i] = fmt.Sprintf("W%d", i)
	}

	planned := make([]int, span)
	for k := 1; k < span; k++ {
		pct := 0.0
		for i := range items {
			pct += items[i].weight * clamp01((float64(k)-items[i].start)/items[i].weeks)
		}
		planned[k] = int(pct * 100)
	}
	planned[span-1] = 100

	actual := make([]int, span)
	monthly := make([]decimal.Decimal, span)
	for i := range items {
		logs, err := h.pool.Query(ctx, `
			SELECT progress, "createdAt" FROM work_item_progress_logs
			WHERE "workItemId" = $1 ORDER BY "createdAt" ASC, id ASC`, items[i].id)
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
			entries = []logEntry{{progress: 0, at: start}}
		}
		prev := 0
		for _, e := range entries {
			if e.progress <= prev {
				prev = e.progress
				continue
			}
			delta := items[i].cost.Mul(decimal.NewFromInt(int64(e.progress - prev))).Div(decimal.NewFromInt(100))
			idx := daysBetween(start, e.at)/7 + 1
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
	cumCost := decimal.Zero
	for i := 1; i < span; i++ {
		cumCost = cumCost.Add(monthly[i])
		pct := int(cumCost.Mul(decimal.NewFromInt(100)).Div(totalCost).IntPart())
		if pct > 100 {
			pct = 100
		}
		actual[i] = pct
	}
	return labels, planned, actual
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func daysBetween(a, b time.Time) int {
	return int(b.Sub(a).Hours() / 24)
}
