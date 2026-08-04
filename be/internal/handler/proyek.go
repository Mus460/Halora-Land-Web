package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

type ProyekHandler struct {
	pool *pgxpool.Pool
	repo *repository.ProyekRepo
}

func NewProyekHandler(pool *pgxpool.Pool, repo *repository.ProyekRepo) *ProyekHandler {
	return &ProyekHandler{pool: pool, repo: repo}
}

func (h *ProyekHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	f := repository.ListProyekFilter{UserID: u.UserID, Search: r.URL.Query().Get("search"), Tipe: r.URL.Query().Get("tipe")}
	if u.Role == models.RoleAdmin {
		f.IsAdmin = true
	}
	out, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"proyek": out})
}

func (h *ProyekHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetDetail(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if !h.canView(r, &models.Proyek{ID: p.ID, UserID: p.UserID}) {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"proyek": p})
}

func (h *ProyekHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		NamaProyek   string           `json:"namaProyek"`
		Lokasi       *string          `json:"lokasi"`
		Tipe         string           `json:"tipe"`
		IsPitching   bool             `json:"isPitching"`
		NilaiKontrak *decimal.Decimal `json:"nilaiKontrak"`
		Timeline     *string          `json:"timeline"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.NamaProyek == "" {
		writeError(w, http.StatusBadRequest, "namaProyek wajib diisi")
		return
	}
	tipe := models.TipeProyekGedung
	if in.Tipe == "infra" {
		tipe = models.TipeProyekInfra
	}
	var nk *decimal.Decimal
	if in.NilaiKontrak != nil && !in.NilaiKontrak.IsZero() {
		nk = in.NilaiKontrak
	}
	p, err := h.repo.Create(r.Context(), repository.CreateProyekInput{
		UserID: u.UserID, NamaProyek: in.NamaProyek, Lokasi: in.Lokasi, Tipe: tipe,
		IsPitching: in.IsPitching, NilaiKontrak: nk, Timeline: in.Timeline,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"proyek": p})
}

func (h *ProyekHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, id, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in repository.UpdateProyekInput
	if !decodeJSON(w, r, &in) {
		return
	}
	updated, err := h.repo.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"proyek": updated})
}

func (h *ProyekHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, id, auth.AccessOwner); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *ProyekHandler) canView(r *http.Request, p *models.Proyek) bool {
	_, _, ok := auth.ProjectAccess(r.Context(), h.pool, p.ID, auth.AccessView)
	return ok
}
