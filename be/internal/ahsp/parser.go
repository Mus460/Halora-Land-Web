package ahsp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// SheetToKategori maps Excel sheet names to kategori slugs (port of
// SHEET_TO_KATEGORI, ARCHITECTURE.md §3.4/§8.1). First 2 sheets are skipped.
var SheetToKategori = map[string]string{
	"Persiapan":       "persiapan",
	"Pondasi":         "pondasi",
	"Beton":           "beton",
	"Kanopi":          "kanopi",
	"Baja":            "baja",
	"Tangga":          "tangga",
	"Atap":            "atap",
	"Dinding":         "dinding",
	"Plesteran":       "plesteran",
	"Acian":           "acian",
	"Keramik":         "keramik",
	"Paving":          "paving",
	"Pengecatan":      "pengecatan",
	"Pintu":           "pintu",
	"Interior":        "interior",
	"Toilet":          "toilet",
	"MEP":             "mep",
}

// WorkItem is a parsed AHSP catalog item ready for master_analisa insertion.
type WorkItem struct {
	Kode        string
	Nama        string
	Satuan      string
	HargaSatuan decimal.Decimal
	BiayaUmum   decimal.Decimal
	Section     string
	Sheet       string
	Breakdown   []BreakdownRow
}

// BreakdownRow is one rincian_analisa template row (§3.2 snapshot-style with
// komponenId=NULL and inline nama/satuan/hargaSatuan/jumlahHarga/kodeReferensi).
type BreakdownRow struct {
	Tipe         string
	Nama         string
	Satuan       string
	Koef         decimal.Decimal
	HargaSatuan  decimal.Decimal
	JumlahHarga  decimal.Decimal
	KodeReferensi string
}

var kodeRe = regexp.MustCompile(`^\d+(\.\d+){3,}$`)
var satuanRe = regexp.MustCompile(`(?i)\b(m1|m2|m3|m|kg|ton|unit|buah|ls|liter|OH|hari|jam)\b`)

// ParseSheet parses one Excel sheet into WorkItems (port of ahsp-parser.ts).
// Work item detection: col 2 = kode (>=4 numeric dot-levels), col 3 = nama,
// col 8 = hargaSatuan. Breakdown: scan up to 100 rows; col 2 section markers
// A->upah, B->material, C->alat, D/E/F->end. Column indices use defensive
// fallbacks (row[4]||row[5]) for non-uniform sheets.
func ParseSheet(f *excelize.File, sheet string) ([]WorkItem, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("read sheet %s: %w", sheet, err)
	}
	kategori := SheetToKategori[sheet]
	if kategori == "" {
		kategori = "custom"
	}

	var items []WorkItem
	for i := 0; i < len(rows); i++ {
		row := padCols(rows[i], 10)
		kode := strings.TrimSpace(row[1])
		if !kodeRe.MatchString(kode) {
			continue
		}
		nama := strings.TrimSpace(row[2])
		if nama == "" {
			continue
		}
		hsStr := strings.TrimSpace(row[7])
		hs, _ := decimal.NewFromString(normalizeNum(hsStr))
		item := WorkItem{
			Kode: kode, Nama: nama, HargaSatuan: hs,
			BiayaUmum: decimal.NewFromFloat(0.10),
			Section: kategori, Sheet: sheet,
			Satuan: extractSatuan(row[3], nama),
		}

		for j := i + 1; j < len(rows) && j < i+100; j++ {
			r := padCols(rows[j], 10)
			marker := strings.TrimSpace(r[1])
			if marker == "" {
				continue
			}
			switch marker {
			case "A":
				item.Breakdown = append(item.Breakdown, parseBreakdown(r, "upah"))
				continue
			case "B":
				item.Breakdown = append(item.Breakdown, parseBreakdown(r, "material"))
				continue
			case "C":
				item.Breakdown = append(item.Breakdown, parseBreakdown(r, "alat"))
				continue
			case "D", "E", "F":
				goto nextItem
			}
			if kodeRe.MatchString(marker) {
				i = j - 1
				goto nextItem
			}
		}
	nextItem:
		items = append(items, item)
	}
	return items, nil
}

func parseBreakdown(row []string, tipe string) BreakdownRow {
	nama := strOr(row, []int{4, 5, 6})
	satuan := strOr(row, []int{5, 6, 4})
	koefStr := strOr(row, []int{6, 5, 7})
	hsStr := strOr(row, []int{7, 8, 6})
	jhStr := strOr(row, []int{8, 9, 7})
	kodeRef := strOr(row, []int{2, 3, 1})
	koef, _ := decimal.NewFromString(normalizeNum(koefStr))
	hs, _ := decimal.NewFromString(normalizeNum(hsStr))
	jh, _ := decimal.NewFromString(normalizeNum(jhStr))
	if !jh.IsZero() && hs.IsZero() && !koef.IsZero() {
		hs = jh.Div(koef)
	}
	return BreakdownRow{
		Tipe: tipe, Nama: nama, Satuan: satuan,
		Koef: koef, HargaSatuan: hs, JumlahHarga: jh, KodeReferensi: kodeRef,
	}
}

func strOr(row []string, idxs []int) string {
	for _, i := range idxs {
		if i < len(row) {
			v := strings.TrimSpace(row[i])
			if v != "" {
				return v
			}
		}
	}
	return ""
}

func padCols(row []string, n int) []string {
	if len(row) >= n {
		return row
	}
	out := make([]string, n)
	copy(out, row)
	return out
}

func normalizeNum(s string) string {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "Rp", "")
	s = strings.ReplaceAll(s, "rp", "")
	return s
}

func extractSatuan(col, nama string) string {
	if s := satuanRe.FindString(col); s != "" {
		return s
	}
	if s := satuanRe.FindString(nama); s != "" {
		return s
	}
	return "unit"
}

// ListSheets returns the importable sheets (skipping the first 2).
func ListSheets(f *excelize.File) []string {
	all := f.GetSheetList()
	if len(all) <= 2 {
		return nil
	}
	out := make([]string, 0, len(all)-2)
	for _, s := range all[2:] {
		out = append(out, s)
	}
	return out
}
