package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/repository"
	"github.com/halora-land/halora-be/internal/service"
)

type ProyekSubHandler struct {
	pool    *pgxpool.Pool
	proyek  *repository.ProyekRepo
	rekap   *repository.RekapRepo
	rab     *service.RABService
	snap    *service.SnapshotService
	real    *repository.RealisasiRepo
	log     *repository.LogistikRepo
	inv     *repository.InvoiceRepo
}

func NewProyekSubHandler(pool *pgxpool.Pool, pr *repository.ProyekRepo, rr *repository.RekapRepo, rab *service.RABService, ss *service.SnapshotService) *ProyekSubHandler {
	return &ProyekSubHandler{
		pool: pool, proyek: pr, rekap: rr, rab: rab, snap: ss,
		real: repository.NewRealisasiRepo(pool),
		log:  repository.NewLogistikRepo(pool),
		inv:  repository.NewInvoiceRepo(pool),
	}
}

func (h *ProyekSubHandler) RekapGet(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	summary, err := h.proyek.Summary(r.Context(), pid)
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

func (h *ProyekSubHandler) RekapPut(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct{ Margin string `json:"margin"` }
	if !decodeJSON(w, r, &in) {
		return
	}
	m, err := decimal.NewFromString(in.Margin)
	if err != nil {
		writeError(w, http.StatusBadRequest, "margin tidak valid")
		return
	}
	if err := h.rekap.UpsertMargin(r.Context(), pid, m); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "margin": m})
}

func (h *ProyekSubHandler) RecalculateAll(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
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

func (h *ProyekSubHandler) RealisasiList(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	out, err := h.real.List(r.Context(), pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProyekSubHandler) LogistikList(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	out, err := h.log.List(r.Context(), pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProyekSubHandler) InvoiceList(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	out, err := h.inv.List(r.Context(), pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}
