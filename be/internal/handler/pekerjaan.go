package handler

import (
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
	"github.com/halora-land/halora-be/internal/service"
)

type PekerjaanHandler struct {
	pool     *pgxpool.Pool
	repo     *repository.PekerjaanRepo
	snapshot *service.SnapshotService
}

func NewPekerjaanHandler(pool *pgxpool.Pool, repo *repository.PekerjaanRepo, ss *service.SnapshotService) *PekerjaanHandler {
	return &PekerjaanHandler{pool: pool, repo: repo, snapshot: ss}
}

func (h *PekerjaanHandler) List(w http.ResponseWriter, r *http.Request) {
	f := repository.ListPekerjaanFilter{Kategori: r.URL.Query().Get("kategori"), Search: r.URL.Query().Get("search")}
	if pid := r.URL.Query().Get("proyekId"); pid != "" {
		if v, err := atoi32(pid); err == nil {
			f.ProyekID = &v
		}
	}
	out, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *PekerjaanHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProyekID, auth.AccessView); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *PekerjaanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ProyekID        int32   `json:"proyekId"`
		Kategori        string  `json:"kategori"`
		UraianPekerjaan string  `json:"uraianPekerjaan"`
		Volume          string  `json:"volume"`
		Satuan          string  `json:"satuan"`
		HargaSatuan     string  `json:"hargaSatuan"`
		MetodeHitung    string  `json:"metodeHitung"`
		TipePekerjaan   *string `json:"tipePekerjaan"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, in.ProyekID, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	vol, _ := decimal.NewFromString(in.Volume)
	hs, _ := decimal.NewFromString(in.HargaSatuan)
	p, err := h.repo.Create(r.Context(), nil, repository.CreatePekerjaanInput{
		ProyekID: in.ProyekID, Kategori: models.KategoriPekerjaan(in.Kategori),
		UraianPekerjaan: in.UraianPekerjaan, Volume: vol, Satuan: in.Satuan,
		HargaSatuan: hs, TotalBiaya: hs.Mul(vol), MetodeHitung: models.MetodeHitung(in.MetodeHitung),
		TipePekerjaan: in.TipePekerjaan,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *PekerjaanHandler) Update(w http.ResponseWriter, r *http.Request) {
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
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProyekID, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct {
		Volume      *string `json:"volume"`
		HargaSatuan *string `json:"hargaSatuan"`
		Uraian      *string `json:"uraianPekerjaan"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	var vol, hs, tb *decimal.Decimal
	if in.Volume != nil {
		d, _ := decimal.NewFromString(*in.Volume)
		vol = &d
	}
	if in.HargaSatuan != nil {
		d, _ := decimal.NewFromString(*in.HargaSatuan)
		hs = &d
	}
	if vol != nil && hs != nil {
		t := hs.Mul(*vol)
		tb = &t
	}
	updated, err := h.repo.Update(r.Context(), id, vol, hs, tb, in.Uraian)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *PekerjaanHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProyekID, auth.AccessOwner); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *PekerjaanHandler) FromAHSP(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct {
		MasterAnalisaID int32  `json:"masterAnalisaId"`
		Volume          string `json:"volume"`
		ApplyBreakdown  *bool  `json:"applyBreakdown"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	vol, err := decimal.NewFromString(in.Volume)
	if err != nil {
		writeError(w, http.StatusBadRequest, "volume tidak valid")
		return
	}
	apply := true
	if in.ApplyBreakdown != nil {
		apply = *in.ApplyBreakdown
	}
	p, err := h.snapshot.FromAHSP(r.Context(), pid, in.MasterAnalisaID, vol, apply)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "pekerjaan": p})
}

func (h *PekerjaanHandler) Recalculate(w http.ResponseWriter, r *http.Request) {
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
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProyekID, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
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

func (h *PekerjaanHandler) ValidateSnapshot(w http.ResponseWriter, r *http.Request) {
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
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, p.ProyekID, auth.AccessView); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	res, err := h.snapshot.ValidateSnapshot(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *PekerjaanHandler) ListAnalisa(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	out, err := h.repo.ListDetailAnalisa(r.Context(), id)
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
