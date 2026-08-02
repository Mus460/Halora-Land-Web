package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

type MasterAnalisaHandler struct {
	pool *pgxpool.Pool
	repo *repository.MasterAnalisaRepo
}

func NewMasterAnalisaHandler(pool *pgxpool.Pool, repo *repository.MasterAnalisaRepo) *MasterAnalisaHandler {
	return &MasterAnalisaHandler{pool: pool, repo: repo}
}

func (h *MasterAnalisaHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	f := repository.ListMasterAnalisaFilter{UserID: u.UserID, Search: r.URL.Query().Get("search")}
	if lvl := r.URL.Query().Get("level"); lvl != "" {
		if v, err := atoi32(lvl); err == nil {
			f.Level = new(int)
			*f.Level = int(v)
		}
	}
	if pid := r.URL.Query().Get("parentId"); pid != "" && pid != "null" {
		if v, err := atoi32(pid); err == nil {
			f.ParentID = &v
		}
	}
	if ig := r.URL.Query().Get("isGlobal"); ig == "true" {
		b := true
		f.IsGlobal = &b
	}

	// When fetching root nodes (level=0, no parentId), build a full tree
	// with nested children so the FE tree component works without N+1 requests.
	if f.Level != nil && *f.Level == 0 && f.ParentID == nil && f.Search == "" {
		tree, err := h.repo.ListTree(r.Context(), f)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": tree})
		return
	}

	out, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}

func (h *MasterAnalisaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	m, err := h.repo.Get(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MasterAnalisaHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		Kode     string  `json:"kode"`
		Nama     string  `json:"nama"`
		Level    int32   `json:"level"`
		ParentID *int32  `json:"parentId"`
		Satuan   *string `json:"satuan"`
		IsGlobal bool    `json:"isGlobal"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.IsGlobal && u.Role != models.RoleAdmin {
		writeError(w, http.StatusForbidden, "Hanya ADMIN yang dapat membuat item global")
		return
	}
	uid := u.UserID
	in_ := repository.CreateMasterAnalisaInput{
		Kode: in.Kode, Nama: in.Nama, Level: in.Level, ParentID: in.ParentID,
		Satuan: in.Satuan, IsGlobal: in.IsGlobal, UserID: &uid,
	}
	m, err := h.repo.Create(r.Context(), in_)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *MasterAnalisaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	has, err := h.repo.HasChildren(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if has {
		writeError(w, http.StatusConflict, "Tidak dapat menghapus item yang memiliki turunan")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *MasterAnalisaHandler) ListRincian(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	out, err := h.repo.ListRincian(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *MasterAnalisaHandler) CreateRincian(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	var in repository.CreateRincianInput
	if !decodeJSON(w, r, &in) {
		return
	}
	in.MasterAnalisaID = id
	if err := h.repo.CreateRincian(r.Context(), in); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true})
}

func (h *MasterAnalisaHandler) DeleteRincian(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	rid, ok := parseIntParam(w, r, "rincianId")
	if !ok {
		return
	}
	if err := h.repo.DeleteRincian(r.Context(), id, rid); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *MasterAnalisaHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	kategori := r.URL.Query().Get("kategori")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := atoi32(l); err == nil && v > 0 {
			limit = int(v)
		}
	}
	out, err := h.repo.SearchAHSP(r.Context(), q, kategori, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": out})
}

type MasterHargaHandler struct {
	pool *pgxpool.Pool
	repo *repository.MasterHargaRepo
}

func NewMasterHargaHandler(pool *pgxpool.Pool, repo *repository.MasterHargaRepo) *MasterHargaHandler {
	return &MasterHargaHandler{pool: pool, repo: repo}
}

func (h *MasterHargaHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	f := repository.ListMasterHargaFilter{UserID: u.UserID, Kategori: r.URL.Query().Get("kategori"), Search: r.URL.Query().Get("search")}
	if ig := r.URL.Query().Get("isGlobal"); ig == "true" {
		b := true
		f.IsGlobal = &b
	}
	out, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"masterHarga": out})
}

func (h *MasterHargaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	m, err := h.repo.Get(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MasterHargaHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		Nama     string          `json:"nama"`
		Satuan   string          `json:"satuan"`
		Harga    decimal.Decimal `json:"harga"`
		Kategori string          `json:"kategori"`
		IsGlobal bool            `json:"isGlobal"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.IsGlobal && u.Role != models.RoleAdmin {
		in.IsGlobal = false
	}
	uid := u.UserID
	m, err := h.repo.Create(r.Context(), repository.CreateMasterHargaInput{
		Nama: in.Nama, Satuan: in.Satuan, Harga: in.Harga,
		Kategori: models.TipeKomponen(in.Kategori), IsGlobal: in.IsGlobal, UserID: &uid,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *MasterHargaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *MasterHargaHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	var in struct {
		Nama     *string          `json:"nama"`
		Satuan   *string          `json:"satuan"`
		Harga    *decimal.Decimal `json:"harga"`
		Kategori *string          `json:"kategori"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	var kat *models.TipeKomponen
	if in.Kategori != nil {
		k := models.TipeKomponen(*in.Kategori)
		kat = &k
	}
	m, err := h.repo.Update(r.Context(), id, repository.UpdateMasterHargaInput{
		Nama: in.Nama, Satuan: in.Satuan, Harga: in.Harga, Kategori: kat,
	})
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
