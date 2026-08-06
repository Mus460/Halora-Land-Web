package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
	"github.com/halora-land/halora-be/service"
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
	var in struct{ Margin decimal.Decimal `json:"margin"` }
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := h.rekap.UpsertMargin(r.Context(), pid, in.Margin); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "margin": in.Margin})
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
	writeJSON(w, http.StatusOK, map[string]any{"realisasi": out})
}

func (h *ProyekSubHandler) RealisasiCreate(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct {
		Tanggal    string               `json:"tanggal"`
		Kategori   string               `json:"kategori"`
		Jumlah     decimal.Decimal      `json:"jumlah"`
		Keterangan *string              `json:"keterangan"`
		Jenis      models.JenisRealisasi `json:"jenis"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Tanggal == "" {
		writeError(w, http.StatusBadRequest, "tanggal wajib diisi")
		return
	}
	if in.Kategori == "" {
		writeError(w, http.StatusBadRequest, "kategori wajib diisi")
		return
	}
	if in.Jumlah.LessThanOrEqual(decimal.Zero) {
		writeError(w, http.StatusBadRequest, "jumlah harus lebih dari 0")
		return
	}
	if in.Jenis == "" {
		in.Jenis = models.RealisasiPengeluaran
	}
	if in.Jenis != models.RealisasiPengeluaran && in.Jenis != models.RealisasiPemasukan {
		writeError(w, http.StatusBadRequest, "jenis tidak valid")
		return
	}
	out, err := h.real.Create(r.Context(), pid, in.Tanggal, in.Kategori, in.Jumlah, in.Keterangan, in.Jenis, models.RealisasiApproved, nil, nil)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"realisasi": out})
}

func (h *ProyekSubHandler) RealisasiApprove(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	rid, ok := parseIntParam(w, r, "rid")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct {
		Status models.StatusRealisasi `json:"status"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Status != models.RealisasiApproved {
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
	writeJSON(w, http.StatusOK, map[string]any{"realisasi": out})
}

func (h *ProyekSubHandler) RealisasiDelete(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	rid, ok := parseIntParam(w, r, "rid")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
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
	writeJSON(w, http.StatusOK, map[string]any{"logistik": out})
}

func (h *ProyekSubHandler) LogistikCreate(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct {
		NamaMaterial  string          `json:"namaMaterial"`
		Satuan        string          `json:"satuan"`
		Volume        decimal.Decimal `json:"volume"`
		HargaSatuan   decimal.Decimal `json:"hargaSatuan"`
		Tanggal       *string         `json:"tanggal"`
		Keterangan    *string         `json:"keterangan"`
		CatatKeuangan bool            `json:"catatKeuangan"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.NamaMaterial == "" {
		writeError(w, http.StatusBadRequest, "nama material wajib diisi")
		return
	}
	if in.Satuan == "" {
		writeError(w, http.StatusBadRequest, "satuan wajib diisi")
		return
	}
	if in.Volume.LessThanOrEqual(decimal.Zero) {
		writeError(w, http.StatusBadRequest, "volume harus lebih dari 0")
		return
	}
	if in.HargaSatuan.IsNegative() {
		writeError(w, http.StatusBadRequest, "harga satuan tidak boleh negatif")
		return
	}
	if in.CatatKeuangan && in.Tanggal == nil {
		today := time.Now().Format("2006-01-02")
		in.Tanggal = &today
	}
	out, linked, err := h.log.Create(r.Context(), pid, in.NamaMaterial, in.Satuan, in.Volume, in.HargaSatuan, in.Tanggal, in.Keterangan, in.CatatKeuangan)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"logistik": out, "realisasi": linked})
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
	writeJSON(w, http.StatusOK, map[string]any{"invoices": out})
}

func (h *ProyekSubHandler) InvoiceCreate(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct {
		Tanggal string               `json:"tanggal"`
		Total   decimal.Decimal      `json:"total"`
		Status  models.StatusInvoice `json:"status"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Tanggal == "" {
		writeError(w, http.StatusBadRequest, "tanggal wajib diisi")
		return
	}
	if in.Total.LessThanOrEqual(decimal.Zero) {
		writeError(w, http.StatusBadRequest, "total harus lebih dari 0")
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
	out, err := h.inv.Create(r.Context(), pid, in.Tanggal, in.Total, in.Status)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"invoice": out})
}

func (h *ProyekSubHandler) InvoiceUpdate(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	iid, ok := parseIntParam(w, r, "iid")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessEdit); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct {
		Status        models.StatusInvoice `json:"status"`
		CatatKeuangan bool                 `json:"catatKeuangan"`
	}
	if !decodeJSON(w, r, &in) {
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
	inv, err := h.inv.UpdateStatus(r.Context(), pid, iid, in.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Invoice tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var linked *models.Realisasi
	if in.Status == models.InvoicePaid && in.CatatKeuangan {
		existing, err := h.real.FindByInvoiceID(r.Context(), pid, iid)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			tanggal := inv.Tanggal.Format("2006-01-02")
			ket := inv.Nomor
			linked, err = h.real.Create(r.Context(), pid, tanggal, "Invoice", inv.Total, &ket, models.RealisasiPemasukan, models.RealisasiDraft, nil, &iid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			linked = existing
		}
	} else if in.Status != models.InvoicePaid {
		existing, err := h.real.FindByInvoiceID(r.Context(), pid, iid)
		if err == nil && existing.Status != models.RealisasiReverted {
			if err := h.real.TagReverted(r.Context(), pid, existing.ID); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			existing.Status = models.RealisasiReverted
			linked = existing
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"invoice": inv, "realisasi": linked})
}

func (h *ProyekSubHandler) KurvaS(w http.ResponseWriter, r *http.Request) {
	pid, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, ok := auth.ProjectAccess(r.Context(), h.pool, pid, auth.AccessView); !ok {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	items, err := h.rab.Compute(r.Context(), pid, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	totalKategori := len(items.Grouped)
	if totalKategori == 0 {
		writeJSON(w, http.StatusOK, map[string]any{
			"labels":  []string{},
			"planned": []int{},
			"actual":  []int{},
		})
		return
	}
	months := []string{"M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9", "M10", "M11", "M12"}
	planned := make([]int, 12)
	for i := range planned {
		planned[i] = (i + 1) * (100 / 12)
	}
	if planned[11] < 100 {
		planned[11] = 100
	}
	actual := make([]int, 12)
	writeJSON(w, http.StatusOK, map[string]any{
		"labels":  months,
		"planned": planned,
		"actual":  actual,
	})
}
