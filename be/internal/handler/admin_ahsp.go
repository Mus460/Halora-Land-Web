package handler

import (
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/halora-land/halora-be/internal/ahsp"
)

type AdminAHSPHandler struct {
	pool     *pgxpool.Pool
	importer *ahsp.Importer
	filePath string
}

func NewAdminAHSPHandler(pool *pgxpool.Pool, importer *ahsp.Importer, filePath string) *AdminAHSPHandler {
	return &AdminAHSPHandler{pool: pool, importer: importer, filePath: filePath}
}

func (h *AdminAHSPHandler) ImportStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.importer.ImportStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *AdminAHSPHandler) Import(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SheetName      string `json:"sheetName"`
		ForceReimport  bool   `json:"forceReimport"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.SheetName == "" {
		writeError(w, http.StatusBadRequest, "sheetName wajib diisi")
		return
	}
	path := h.filePath
	if p := os.Getenv("AHSP_XLSX_PATH"); p != "" {
		path = p
	}
	f, err := openXlsx(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal membuka xlsx: "+err.Error())
		return
	}
	defer f.Close()
	items, err := ahsp.ParseSheet(f, in.SheetName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.importer.ImportSheet(r.Context(), items, in.ForceReimport)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}
