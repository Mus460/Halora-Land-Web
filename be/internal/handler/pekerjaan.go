package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
	"github.com/halora-land/halora-be/service"
)

type WorkItemHandler struct {
	pool     database.Pool
	repo     *repository.WorkItemRepo
	snapshot *service.SnapshotService
	progress *service.ProgressService
}

func NewWorkItemHandler(pool database.Pool, repo *repository.WorkItemRepo, ss *service.SnapshotService, progress *service.ProgressService) *WorkItemHandler {
	return &WorkItemHandler{pool: pool, repo: repo, snapshot: ss, progress: progress}
}

func (h *WorkItemHandler) List(w http.ResponseWriter, r *http.Request) {
	f := repository.ListWorkItemFilter{Category: r.URL.Query().Get("category"), Search: r.URL.Query().Get("search")}
	if pid := r.URL.Query().Get("projectId"); pid != "" {
		if v, err := atoi32(pid); err == nil {
			f.ProjectID = &v
		}
	}
	out, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if f.ProjectID != nil {
		items, _, err := h.progress.Items(r.Context(), *f.ProjectID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		weights := make(map[int32]float64, len(items))
		for i := range items {
			weights[items[i].ID] = items[i].Weight
		}
		for i := range out {
			out[i].Weight = weights[out[i].ID]
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *WorkItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.Get(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProjectID, auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *WorkItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ProjectID         int32           `json:"projectId"`
		Category          string          `json:"category"`
		Description       string          `json:"description"`
		Volume            decimal.Decimal `json:"volume"`
		Unit              string          `json:"unit"`
		UnitPrice         decimal.Decimal `json:"unitPrice"`
		CalculationMethod string          `json:"calculationMethod"`
		Level             *string         `json:"level"`
		Type              *string         `json:"type"`
		AnalysisMasterID  *int32          `json:"analysisMasterId"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, in.ProjectID, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	if in.AnalysisMasterID != nil {
		kat := models.WorkCategory(in.Category)
		p, err := h.snapshot.FromAHSP(r.Context(), in.ProjectID, *in.AnalysisMasterID, in.Volume, true, &kat)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, p)
		return
	}
	var duration *decimal.Decimal
	if in.AnalysisMasterID != nil {
		wk, err := h.repo.DurationCoefficient(r.Context(), *in.AnalysisMasterID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		duration = wk
	}
	p, err := h.repo.Create(r.Context(), nil, repository.CreateWorkItemInput{
		ProjectID: in.ProjectID, Category: models.WorkCategory(in.Category),
		Description: in.Description, Volume: in.Volume, Unit: in.Unit,
		UnitPrice: in.UnitPrice, TotalCost: in.UnitPrice.Mul(in.Volume), CalculationMethod: models.CalculationMethod(in.CalculationMethod),
		Level: in.Level, Type: in.Type,
		AnalysisMasterID: in.AnalysisMasterID, Duration: duration,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *WorkItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProjectID, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		Volume            *decimal.Decimal `json:"volume"`
		UnitPrice         *decimal.Decimal `json:"unitPrice"`
		Description       *string          `json:"description"`
		Unit              *string          `json:"unit"`
		Level             *string          `json:"level"`
		Type              *string          `json:"type"`
		CalculationMethod *string          `json:"calculationMethod"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	var tb *decimal.Decimal
	if in.Volume != nil && in.UnitPrice != nil {
		t := in.UnitPrice.Mul(*in.Volume)
		tb = &t
	}
	var mh *models.CalculationMethod
	if in.CalculationMethod != nil {
		m := models.CalculationMethod(*in.CalculationMethod)
		mh = &m
	}
	updated, err := h.repo.Update(r.Context(), id, repository.UpdateWorkItemInput{
		Volume: in.Volume, UnitPrice: in.UnitPrice, TotalCost: tb,
		Description: in.Description, Unit: in.Unit, Level: in.Level,
		Type: in.Type, CalculationMethod: mh,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *WorkItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProjectID, auth.AccessOwner); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *WorkItemHandler) FromAHSP(w http.ResponseWriter, r *http.Request) {
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
		AnalysisMasterID int32           `json:"analysisMasterId"`
		Volume           decimal.Decimal `json:"volume"`
		ApplyBreakdown   *bool           `json:"applyBreakdown"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	apply := true
	if in.ApplyBreakdown != nil {
		apply = *in.ApplyBreakdown
	}
	p, err := h.snapshot.FromAHSP(r.Context(), pid, in.AnalysisMasterID, in.Volume, apply, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "work_items": p})
}

// UpdateDetails bulk-replaces a work item's local breakdown rows (per-project
// customization). Recomputed flat: unitPrice = Σ (coefficient × unitPrice),
// totalCost = unitPrice × volume.
func (h *WorkItemHandler) UpdateDetails(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProjectID, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		Details []detailRow `json:"details"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if len(in.Details) == 0 {
		writeError(w, http.StatusBadRequest, "details minimal 1 item")
		return
	}
	for _, d := range in.Details {
		if d.Name == "" {
			writeError(w, http.StatusBadRequest, "nama komponen wajib diisi")
			return
		}
		if d.Coefficient.IsNegative() || d.UnitPrice.IsNegative() {
			writeError(w, http.StatusBadRequest, "koefisien dan harga tidak boleh negatif")
			return
		}
		if d.Type == "" {
			writeError(w, http.StatusBadRequest, "tipe komponen wajib diisi")
			return
		}
	}

	tx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	var keepIDs []int32
	var sum decimal.Decimal
	for _, d := range in.Details {
		rowTotal := d.Coefficient.Mul(d.UnitPrice)
		sum = sum.Add(rowTotal)
		if d.ID > 0 {
			if err := h.repo.UpdateDetail(r.Context(), tx, repository.UpdateDetailInput{
				ID: d.ID, WorkItemID: id, Name: d.Name, Unit: d.Unit,
				Coefficient: d.Coefficient, UnitPrice: d.UnitPrice, TotalCost: rowTotal, Type: models.ComponentType(d.Type),
			}); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			keepIDs = append(keepIDs, d.ID)
			continue
		}
		if err := h.repo.CreateDetail(r.Context(), tx, repository.CreateDetailInput{
			WorkItemID:  id,
			Name:        d.Name,
			Unit:        d.Unit,
			Coefficient: d.Coefficient,
			UnitPrice:   d.UnitPrice,
			TotalCost:   rowTotal,
			Type:        models.ComponentType(d.Type),
			SourceCode:  d.SourceCode,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := h.repo.DeleteDetailsNotIn(r.Context(), tx, id, keepIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.repo.SetTotal(r.Context(), tx, id, sum, sum.Mul(p.Volume)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := h.repo.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

type detailRow struct {
	ID          int32           `json:"id"`
	Name        string          `json:"name"`
	Unit        string          `json:"unit"`
	Coefficient decimal.Decimal `json:"coefficient"`
	UnitPrice   decimal.Decimal `json:"unitPrice"`
	Type        string          `json:"type"`
	SourceCode  *string         `json:"sourceCode"`
}

func (h *WorkItemHandler) Recalculate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProjectID, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	u := auth.FromContext(r.Context())
	out, err := h.snapshot.Recalculate(r.Context(), id, u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *WorkItemHandler) ValidateSnapshot(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProjectID, auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	res, err := h.snapshot.ValidateSnapshot(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *WorkItemHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProjectID, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in struct {
		Progress *int    `json:"progress"`
		Note     *string `json:"note"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Progress == nil {
		writeError(w, http.StatusBadRequest, "progress wajib diisi")
		return
	}
	log, err := h.repo.RecordProgress(r.Context(), id, *in.Progress, in.Note)
	if err != nil {
		if errors.Is(err, repository.ErrProgressNotIncreasing) {
			writeError(w, http.StatusBadRequest, repository.ErrProgressNotIncreasing.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "progress": clampProgress(*in.Progress), "log": log})
}

func (h *WorkItemHandler) ProgressLogs(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProjectID, auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	out, err := h.repo.ListProgressLogs(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func clampProgress(p int) int {
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

func (h *WorkItemHandler) ListAnalisa(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	out, err := h.repo.ListWorkItemDetail(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func atoi32(s string) (int32, error) {
	v, err := strconv.Atoi(s)
	return int32(v), err
}
