package ahsp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// SheetToCategory maps Excel sheet names (AHSP DJBK-No-47-Tahun-2026
// Bidang Cipta Karya, file ahsp-2026.xlsx) to category slugs. The first two
// sheets (index + price list) are not work-item sheets.
var SheetToCategory = map[string]string{
	"Persiapan":                      "persiapan",
	"Galian Tanah":                   "galian-tanah",
	"Timbunan Pemadatan":             "timbunan-pemadatan",
	"Angkut Material":                "angkut-material",
	"Geotekstil & Geomembran":        "geotekstil-geomembran",
	"Pembongkaran":                   "pembongkaran",
	"Rangka Atap":                    "rangka-atap",
	"Beton":                          "beton",
	"Pondasi":                        "pondasi",
	"BAJA":                           "baja",
	"Beton Pracetak":                 "beton-pracetak",
	"Beton Prategang":                "beton-prategang",
	"Struktur Kayu":                  "struktur-kayu",
	"Dinding Penahan Tanah":          "dinding-penahan-tanah",
	"Penutup Atap":                   "penutup-atap",
	"Plafon":                         "plafon",
	"Pasangan Dinding":               "pasangan-dinding",
	"Plesteran Dan Acian":            "plesteran-acian",
	"Pengecatan dan Pelituran":       "pengecatan-pelituran",
	"Penutup Lantai dan Dinding":     "penutup-lantai-dinding",
	"Pintu dan Jendela":              "pintu-jendela",
	"Kaca":                           "kaca",
	"Besi dan Aluminium":             "besi-aluminium",
	"Kayu":                           "kayu",
	"Ornamen":                        "ornamen",
	"Signage":                        "signage",
	"Sanitair":                       "sanitair",
	"Lansekap":                       "lansekap",
	"Jaringan Listrik":               "jaringan-listrik",
	"Perpipaan dan proteksi kebakar": "perpipaan-proteksi-kebakaran",
	"Sistem Air Minum":               "sistem-air-minum",
	"Sistem Air Limbah":              "sistem-air-limbah",
	"Bak Kontrol":                    "bak-kontrol",
	"Perpipaan dalam Gedung":         "perpipaan-dalam-gedung",
	"Sistem Air Hujan":               "sistem-air-hujan",
	"Jalan Pada Pemukiman":           "jalan-pemukiman",
	"Drainase":                       "drainase",
	"Jaringan Pipa Luar Gedung":      "jaringan-pipa-luar-gedung",
	"Strutur Risha":                  "struktur-risha",
}

// WorkItem is a parsed AHSP catalog item ready for analysis_masters insertion.
type WorkItem struct {
	Code        string
	Name        string
	Unit        string
	UnitPrice   decimal.Decimal
	GeneralCost decimal.Decimal
	Section     string
	Sheet       string
	Breakdown   []BreakdownRow
}

// BreakdownRow is one analysis_components template row (snapshot-style with
// componentId=NULL and inline name/unit/unitPrice/totalPrice/referenceCode).
type BreakdownRow struct {
	Type          string
	Name          string
	Unit          string
	Coefficient   decimal.Decimal
	UnitPrice     decimal.Decimal
	TotalPrice    decimal.Decimal
	ReferenceCode string
}

// PriceItem is one row of the "Upah & Bahan" price list sheet.
type PriceItem struct {
	Code  string
	Name  string
	Unit  string
	Price decimal.Decimal
	Type  string
}

var kodeRe = regexp.MustCompile(`^\d+(\.\d+)+$`)
var rowNoRe = regexp.MustCompile(`^\d+$`)
var sectionRe = regexp.MustCompile(`^(I|II|III)\.?$`)
var unitRe = regexp.MustCompile(`(?i)\b(m1|m2|m3|m|m'|m"|kg|ton|btg|bh|unit|buah|set|ls|liter|OH|OJ|hari|jam)\b`)

const (
	typeLabor     = "upah"
	typeMaterial  = "material"
	typeEquipment = "alat"
)

// sheetLayout describes the per-item breakdown column positions, detected
// from each item's own header row ("No | Description | Code | Sat. | Coefficient...").
type sheetLayout struct {
	no, description, kodeRef, unit, coefficient int
}

// ParseSheet parses one Excel sheet into WorkItems. Item code lives in column
// C (index 2); each item is followed by a header row and A/B/C sections.
func ParseSheet(f *excelize.File, sheet string) ([]WorkItem, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("read sheet %s: %w", sheet, err)
	}
	category := SheetToCategory[sheet]
	if category == "" {
		category = slugify(sheet)
	}
	unitIndex := ParseIndex(f)

	codeCol := detectCodeCol(rows)
	var items []WorkItem
	for i := 0; i < len(rows); i++ {
		row := padCols(rows[i], 12)
		code := strings.TrimSpace(row[codeCol])
		if !kodeRe.MatchString(code) {
			continue
		}
		name := strings.TrimSpace(row[3])
		if name == "" {
			continue
		}

		hdr := -1
		j := i + 1
		for ; j < len(rows); j++ {
			c := strings.TrimSpace(padCols(rows[j], 12)[codeCol])
			if kodeRe.MatchString(c) {
				break
			}
			if c == "No" && strings.TrimSpace(padCols(rows[j], 12)[3]) == "Uraian" {
				hdr = j
				break
			}
			if c == "A" || c == "B" || c == "C" {
				hdr = j
				break
			}
		}
		if hdr < 0 {
			continue // group title without breakdown
		}

		lay := defaultLayout(codeCol)
		if strings.TrimSpace(padCols(rows[hdr], 12)[codeCol]) == "No" {
			lay = detectLayout(rows[hdr])
		}

		item := WorkItem{
			Code: code, Name: name, GeneralCost: decimal.NewFromFloat(0.10),
			Section: category, Sheet: sheet,
			Unit: unitIndex[code],
		}
		if item.Unit == "" {
			item.Unit = extractUnit(name)
		}

		curType := ""
		j = hdr + 1
		for ; j < len(rows); j++ {
			r := padCols(rows[j], 12)
			c := strings.TrimSpace(r[codeCol])
			if t := sectionType(c); t != "" {
				curType = t
				continue
			}
			if c == "D" || c == "E" || c == "F" {
				break
			}
			if kodeRe.MatchString(c) {
				break // next item
			}
			if strings.HasPrefix(strings.ToUpper(c), "JUMLAH") || curType == "" {
				continue
			}
			if !rowNoRe.MatchString(c) {
				continue // subsection markers like "B.1"
			}
			brow := strings.TrimSpace(r[lay.description])
			if brow == "" {
				continue
			}
			coefficient, _ := decimal.NewFromString(normalizeNum(r[lay.coefficient]))
			b := BreakdownRow{Type: curType, Name: brow, Unit: strings.TrimSpace(r[lay.unit]), Coefficient: coefficient}
			if lay.kodeRef >= 0 {
				b.ReferenceCode = strings.TrimSpace(r[lay.kodeRef])
			}
			item.Breakdown = append(item.Breakdown, b)
		}
		i = j - 1
		items = append(items, item)
	}
	return items, nil
}

func sectionType(marker string) string {
	switch marker {
	case "A":
		return typeLabor
	case "B":
		return typeMaterial
	case "C":
		return typeEquipment
	}
	return ""
}

// ParseIndex reads the "Daftar Harga Satuan Pekerjaan" sheet (code -> unit).
func ParseIndex(f *excelize.File) map[string]string {
	out := map[string]string{}
	rows, err := f.GetRows("Daftar Harga Satuan Pekerjaan")
	if err != nil {
		return out
	}
	for _, row := range rows {
		if len(row) < 4 {
			continue
		}
		code := normalizeKode(strings.TrimSpace(row[1]))
		if code == "" {
			continue
		}
		if unit := strings.TrimSpace(row[3]); unit != "" {
			out[code] = unit
		}
	}
	return out
}

func normalizeKode(k string) string {
	if k == "" {
		return k
	}
	if strings.Count(k, ".") == 1 {
		k = strings.TrimRight(strings.TrimRight(k, "0"), ".")
	}
	return k
}

// ParsePriceList reads the "Upah & Bahan" sheet (sections I./II./III. =>
// upah/material/alat). Rows without a numeric price are skipped.
func ParsePriceList(f *excelize.File) ([]PriceItem, error) {
	const sheet = "Upah & Bahan"
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("read sheet %s: %w", sheet, err)
	}
	hdr := -1
	for i, row := range rows {
		if len(row) > 7 && strings.EqualFold(strings.TrimSpace(row[7]), "HARGA SATUAN") {
			hdr = i
			break
		}
	}
	if hdr < 0 {
		return nil, fmt.Errorf("kolom HARGA SATUAN tidak ditemukan di sheet %s", sheet)
	}

	curType := ""
	var out []PriceItem
	for i := hdr + 1; i < len(rows); i++ {
		r := padCols(rows[i], 9)
		if sectionRe.MatchString(strings.TrimSpace(r[3])) {
			switch strings.TrimSpace(r[3]) {
			case "I.", "I":
				curType = typeLabor
			case "II.", "II":
				curType = typeMaterial
			default:
				curType = typeEquipment
			}
			continue
		}
		name := strings.TrimSpace(r[5])
		if name == "" || curType == "" {
			continue
		}
		h, err := decimal.NewFromString(normalizeNum(strings.TrimSpace(r[7])))
		if err != nil || h.IsZero() {
			continue
		}
		unit := strings.TrimSpace(r[6])
		if unit == "" {
			unit = "unit"
		}
		out = append(out, PriceItem{
			Code: strings.TrimSpace(r[4]), Name: name, Unit: unit, Price: h, Type: curType,
		})
	}
	return out, nil
}

func detectCodeCol(rows [][]string) int {
	count := func(col int) int {
		n := 0
		limit := len(rows)
		if limit > 30 {
			limit = 30
		}
		for _, row := range rows[:limit] {
			if len(row) > col && kodeRe.MatchString(strings.TrimSpace(row[col])) {
				n++
			}
		}
		return n
	}
	if count(2) < count(1) {
		return 1
	}
	return 2
}

func defaultLayout(codeCol int) sheetLayout {
	return sheetLayout{no: codeCol, description: 3, kodeRef: 4, unit: 5, coefficient: 6}
}

func detectLayout(header []string) sheetLayout {
	lay := defaultLayout(2)
	lay.kodeRef = -1
	for i, v := range header {
		t := strings.ToLower(strings.TrimSpace(v))
		switch {
		case strings.HasPrefix(t, "koef"):
			lay.coefficient = i
		case strings.HasPrefix(t, "sat"):
			lay.unit = i
		case t == "kode":
			lay.kodeRef = i
		case t == "uraian":
			lay.description = i
		case t == "no":
			lay.no = i
		}
	}
	return lay
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

func extractUnit(name string) string {
	if s := unitRe.FindString(name); s != "" {
		return s
	}
	return "unit"
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
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
