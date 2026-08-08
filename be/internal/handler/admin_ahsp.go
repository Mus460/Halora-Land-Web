package handler

import (
	"net/http"
	"os"
	"sort"

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
	statusMap, err := h.importer.ImportStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	path := h.filePath
	if p := os.Getenv("AHSP_XLSX_PATH"); p != "" {
		path = p
	}

	var sheets []string
	if f, err := openXlsx(path); err == nil {
		sheets = ahsp.ListSheets(f)
		f.Close()
	}
	if len(sheets) == 0 {
		for sheet := range ahsp.SheetToCategory {
			sheets = append(sheets, sheet)
		}
		sort.Strings(sheets)
	}

	type statusItem struct {
		SheetName string `json:"sheetName"`
		Category  string `json:"category"`
		Imported  bool   `json:"imported"`
		Count     int    `json:"count"`
	}

	out := make([]statusItem, 0, len(sheets))
	for _, sheet := range sheets {
		s := statusItem{
			SheetName: sheet,
			Category:  ahsp.SheetToCategory[sheet],
		}
		if st, ok := statusMap[sheet]; ok {
			s.Imported = st.Imported
			s.Count = st.Count
		}
		out = append(out, s)
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": out})
}

func (h *AdminAHSPHandler) Import(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SheetName     string `json:"sheetName"`
		ForceReimport bool   `json:"forceReimport"`
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
	prices, err := ahsp.ParsePriceList(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if _, err := h.importer.ImportPriceList(r.Context(), prices, in.ForceReimport); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	items, err := ahsp.ParseSheet(f, in.SheetName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.importer.ImportSheet(r.Context(), items, prices, in.ForceReimport)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"imported":   res.Items,
		"skipped":    res.Skipped,
		"count":      res.Items + res.Skipped,
		"components": res.Components,
		"sheet":      res.Sheet,
	})
}
