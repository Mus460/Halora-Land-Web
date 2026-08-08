package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

type AnalysisMasterHandler struct {
	pool *pgxpool.Pool
	repo *repository.AnalysisMasterRepo
}

func NewAnalysisMasterHandler(pool *pgxpool.Pool, repo *repository.AnalysisMasterRepo) *AnalysisMasterHandler {
	return &AnalysisMasterHandler{pool: pool, repo: repo}
}

func (h *AnalysisMasterHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	f := repository.ListAnalysisMasterFilter{UserID: u.UserID, Search: r.URL.Query().Get("search")}
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

func (h *AnalysisMasterHandler) Get(w http.ResponseWriter, r *http.Request) {
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

func (h *AnalysisMasterHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Level    int32   `json:"level"`
		ParentID *int32  `json:"parentId"`
		Unit     *string `json:"unit"`
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
	in_ := repository.CreateAnalysisMasterInput{
		Code: in.Code, Name: in.Name, Level: in.Level, ParentID: in.ParentID,
		Unit: in.Unit, IsGlobal: in.IsGlobal, UserID: &uid,
	}
	m, err := h.repo.Create(r.Context(), in_)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *AnalysisMasterHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func (h *AnalysisMasterHandler) ListComponents(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	out, err := h.repo.ListComponents(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *AnalysisMasterHandler) CreateComponent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	var in repository.CreateComponentInput
	if !decodeJSON(w, r, &in) {
		return
	}
	in.AnalysisMasterID = id
	if err := h.repo.CreateComponent(r.Context(), in); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true})
}

func (h *AnalysisMasterHandler) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	rid, ok := parseIntParam(w, r, "componentId")
	if !ok {
		return
	}
	if err := h.repo.DeleteComponent(r.Context(), id, rid); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// Copy duplicates an analysis master into a user-owned, editable row.
func (h *AnalysisMasterHandler) Copy(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	u := auth.FromContext(r.Context())
	var in struct {
		Name *string `json:"name"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	name := ""
	if in.Name != nil {
		name = *in.Name
	}
	m, err := h.repo.Copy(r.Context(), id, u.UserID, name)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *AnalysisMasterHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	u := auth.FromContext(r.Context())
	var in struct {
		Name *string `json:"name"`
		Unit *string `json:"unit"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	m, err := h.repo.Update(r.Context(), id, u.UserID, u.Role == models.RoleAdmin, repository.UpdateAnalysisMasterInput{Name: in.Name, Unit: in.Unit})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusForbidden, "Tidak dapat mengubah data AHSP sistem")
			return
		}
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *AnalysisMasterHandler) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	cid, ok := parseIntParam(w, r, "componentId")
	if !ok {
		return
	}
	u := auth.FromContext(r.Context())
	var in struct {
		Coefficient *decimal.Decimal `json:"coefficient"`
		UnitPrice   *decimal.Decimal `json:"unitPrice"`
		Unit        *string          `json:"unit"`
		Name        *string          `json:"name"`
		Type        *string          `json:"type"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	var ct *models.ComponentType
	if in.Type != nil {
		t := models.ComponentType(*in.Type)
		ct = &t
	}
	comp, err := h.repo.UpdateComponent(r.Context(), u.UserID, u.Role == models.RoleAdmin, repository.UpdateComponentInput{
		ID: cid, AnalysisMasterID: id,
		Coefficient: in.Coefficient, UnitPrice: in.UnitPrice, Unit: in.Unit, Name: in.Name, Type: ct,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusForbidden, "Tidak dapat mengubah data AHSP sistem")
			return
		}
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusOK, comp)
}

func (h *AnalysisMasterHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("type")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := atoi32(l); err == nil && v > 0 {
			limit = int(v)
		}
	}
	out, err := h.repo.SearchAHSP(r.Context(), q, category, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": out})
}

type PriceMasterHandler struct {
	pool *pgxpool.Pool
	repo *repository.PriceMasterRepo
}

func NewPriceMasterHandler(pool *pgxpool.Pool, repo *repository.PriceMasterRepo) *PriceMasterHandler {
	return &PriceMasterHandler{pool: pool, repo: repo}
}

func (h *PriceMasterHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	f := repository.ListPriceMasterFilter{UserID: u.UserID, Type: r.URL.Query().Get("type"), Search: r.URL.Query().Get("search")}
	if ig := r.URL.Query().Get("isGlobal"); ig == "true" {
		b := true
		f.IsGlobal = &b
	}
	out, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"priceMaster": out})
}

func (h *PriceMasterHandler) Get(w http.ResponseWriter, r *http.Request) {
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

func (h *PriceMasterHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		Name     string          `json:"name"`
		Unit     string          `json:"unit"`
		Price    decimal.Decimal `json:"price"`
		Type     string          `json:"type"`
		IsGlobal bool            `json:"isGlobal"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.IsGlobal && u.Role != models.RoleAdmin {
		in.IsGlobal = false
	}
	uid := u.UserID
	m, err := h.repo.Create(r.Context(), repository.CreatePriceMasterInput{
		Name: in.Name, Unit: in.Unit, Price: in.Price,
		Type: models.ComponentType(in.Type), IsGlobal: in.IsGlobal, UserID: &uid,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *PriceMasterHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func (h *PriceMasterHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	var in struct {
		Name  *string          `json:"name"`
		Unit  *string          `json:"unit"`
		Price *decimal.Decimal `json:"price"`
		Type  *string          `json:"type"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	var kat *models.ComponentType
	if in.Type != nil {
		k := models.ComponentType(*in.Type)
		kat = &k
	}
	m, err := h.repo.Update(r.Context(), id, repository.UpdatePriceMasterInput{
		Name: in.Name, Unit: in.Unit, Price: in.Price, Type: kat,
	})
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
