package boq

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"

	"github.com/halora-land/halora-be/internal/models"
)

// DivisionCategory maps the BOQ division letters to the app's work categories.
// NOTE: this is a temporary fixed map (based on the sample RUKO 2 LANTAI file)
// until divisions become a fixed master list.
var DivisionCategory = map[string]models.WorkCategory{
	"A": models.CategoryPreparation,
	"B": models.CategoryFoundation,
	"C": models.CategoryWall,
	"D": models.CategoryTiles,
	"E": models.CategoryRoof,
	"F": models.CategoryDoors,
	"G": models.CategoryToilet,
	"H": models.CategoryCustom,
}

// DivisionName is the human-readable name of each division letter.
var DivisionName = map[string]string{
	"A": "Persiapan",
	"B": "Struktur & Pondasi",
	"C": "Dinding & Plesteran",
	"D": "Lantai & Plafon",
	"E": "Atap",
	"F": "Pintu & Kusen",
	"G": "Sanitair",
	"H": "Custom",
}

// Item is one BOQ/RAB row ready for work_items insertion.
type Item struct {
	Division    string
	Description string
	Unit        string
	Volume      decimal.Decimal
	UnitPrice   decimal.Decimal
	TotalCost   decimal.Decimal
	Category    models.WorkCategory
}

// Division is one row of the REKAP PER DIVISI section.
type Division struct {
	Letter      string
	Name        string
	Amount      decimal.Decimal
	Description string
}

// Document is the parsed BOQ file.
type Document struct {
	Title     string
	Items     []Item
	Divisions []Division
	Subtotal  decimal.Decimal
	PPN       decimal.Decimal
	Total     decimal.Decimal
}

var divRe = regexp.MustCompile(`^[A-H]$`)
var unitAlias = map[string]string{
	"m'": "m1", "m1": "m1", "m²": "m2", "m2": "m2", "m³": "m3", "m3": "m3",
	"buah": "unit", "bh": "unit", "unit": "unit", "ls": "ls",
}

// Parse reads a BOQ & RAB xlsx into a Document. The single "BOQ & RAB" sheet
// holds: a title row, a header row (NO DIVISI URAIAN PEKERJAAN SATUAN VOLUME
// HARGA SATUAN JUMLAH), item rows, a SUB TOTAL / PPN / TOTAL RAB block and a
// REKAP PER DIVISI section.
func Parse(f *excelize.File) (*Document, error) {
	var sheet string
	for _, s := range f.GetSheetList() {
		if strings.Contains(strings.ToLower(s), "boq") {
			sheet = s
			break
		}
	}
	if sheet == "" {
		all := f.GetSheetList()
		if len(all) == 0 {
			return nil, fmt.Errorf("boq: file tidak memiliki sheet")
		}
		sheet = all[0]
	}
	rows, err := f.GetRows(sheet, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("boq: baca sheet %s: %w", sheet, err)
	}

	doc := &Document{}
	for _, row := range rows {
		if len(row) > 0 && strings.TrimSpace(row[0]) != "" {
			doc.Title = strings.TrimSpace(row[0])
			break
		}
	}

	// Header detection: row that names the "URAIAN PEKERJAAN" column.
	hdr := -1
	for i, row := range rows {
		joined := ""
		for _, c := range row {
			joined += " " + c
		}
		if strings.Contains(strings.ToUpper(joined), "URAIAN") &&
			(strings.Contains(strings.ToUpper(joined), "VOLUME") || strings.Contains(strings.ToUpper(joined), "SATUAN")) {
			hdr = i
			break
		}
	}
	if hdr < 0 {
		return nil, fmt.Errorf("boq: header URAIAN/VOLUME tidak ditemukan di sheet %s", sheet)
	}

	// Item rows: NO DIVISI URAIAN SATUAN VOLUME HARGA JUMLAH
	for i := hdr + 1; i < len(rows); i++ {
		row := pad(rows[i], 7)
		div := strings.TrimSpace(row[1])
		if !divRe.MatchString(div) {
			continue
		}
		desc := strings.TrimSpace(row[2])
		if desc == "" {
			continue
		}
		vol, err := parseNum(row[4])
		if err != nil {
			continue
		}
		hs, err := parseNum(row[5])
		if err != nil {
			continue
		}
		jml, _ := parseNum(row[6])
		if jml.IsZero() {
			jml = vol.Mul(hs).Round(2)
		}
		unit := normalizeUnit(row[3])
		cat, ok := DivisionCategory[div]
		if !ok {
			cat = models.CategoryCustom
		}
		doc.Items = append(doc.Items, Item{
			Division: div, Description: desc, Unit: unit,
			Volume: vol, UnitPrice: hs, TotalCost: jml, Category: cat,
		})
	}

	// Totals block (SUB TOTAL / PPN / TOTAL RAB) and REKAP PER DIVISI.
	inRekap := false
	for _, row := range rows {
		joined := ""
		for _, c := range row {
			joined += " " + strings.TrimSpace(c)
		}
		up := strings.ToUpper(joined)
		switch {
		case strings.Contains(up, "SUB TOTAL"):
			if n, err := lastNum(row); err == nil {
				doc.Subtotal = n
			}
		case strings.Contains(up, "TOTAL RAB"):
			if n, err := lastNum(row); err == nil {
				doc.Total = n
			}
		case strings.Contains(up, "PPN"):
			if n, err := lastNum(row); err == nil {
				doc.PPN = n
			}
		case strings.Contains(up, "REKAP PER DIVISI"):
			inRekap = true
		case inRekap && strings.Contains(up, "DIVISI"):
			// header of the rekap table — nothing to parse
		case inRekap:
			row = pad(row, 2)
			div := strings.TrimSpace(row[0])
			if divRe.MatchString(div) {
				if n, err := parseNum(row[1]); err == nil {
					doc.Divisions = append(doc.Divisions, Division{
						Letter: div, Name: DivisionName[div], Amount: n, Description: DivisionName[div],
					})
				}
			}
			if strings.Contains(up, "CATATAN") {
				inRekap = false
			}
		}
	}

	if len(doc.Items) == 0 {
		return nil, fmt.Errorf("boq: tidak ada item pekerjaan ditemukan di sheet %s", sheet)
	}
	return doc, nil
}

func lastNum(row []string) (decimal.Decimal, error) {
	for i := len(row) - 1; i >= 0; i-- {
		if strings.TrimSpace(row[i]) == "" {
			continue
		}
		return parseNum(row[i])
	}
	return decimal.Zero, fmt.Errorf("no number")
}

// parseNum parses a numeric cell. Number cells arrive as plain dot-decimal
// (RawCellValue), while text cells use Indonesian conventions: comma as the
// decimal separator ("18,5") and dots as thousands separators ("1.800.000").
// "Rp" prefixes are stripped.
func parseNum(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "Rp", "")
	s = strings.ReplaceAll(s, "rp", "")
	s = strings.TrimSpace(s)
	if s == "" {
		return decimal.Zero, fmt.Errorf("empty")
	}
	// Comma is always the decimal separator in Indonesian text cells.
	// Split integer/fraction: "1.800.000,50" → "1800000" + "50", "2,625" → 2.625.
	if i := strings.Index(s, ","); i >= 0 {
		intPart := strings.ReplaceAll(s[:i], ".", "")
		fracPart := s[i+1:]
		s = intPart + "." + fracPart
	} else {
		// Multiple dots mean thousands ("1.000.000"). A single dot is a decimal
		// ("60.5") unless it groups exactly three digits, e.g. "15.000" → 15000.
		switch strings.Count(s, ".") {
		case 0:
			// plain integer
		case 1:
			if len(s)-strings.Index(s, ".")-1 == 3 {
				s = strings.ReplaceAll(s, ".", "")
			}
		default:
			s = strings.ReplaceAll(s, ".", "")
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(f), nil
}

func normalizeUnit(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if u, ok := unitAlias[s]; ok {
		return u
	}
	if s == "" {
		return "unit"
	}
	return s
}

func pad(row []string, n int) []string {
	if len(row) >= n {
		return row
	}
	out := make([]string, n)
	copy(out, row)
	return out
}
