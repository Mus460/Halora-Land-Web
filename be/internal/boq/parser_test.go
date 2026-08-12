package boq

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// buildDoc creates an in-memory BOQ & RAB workbook shaped like the sample
// RUKO 2 LANTAI file.
func buildDoc(t *testing.T, rows [][]string) *excelize.File {
	t.Helper()
	f := excelize.NewFile()
	sheet := "BOQ & RAB"
	f.SetSheetName("Sheet1", sheet)
	for i, row := range rows {
		for j, cell := range row {
			col := string(rune('A' + j))
			if err := f.SetCellValue(sheet, col+itoa(i+1), cell); err != nil {
				t.Fatalf("SetCellValue: %v", err)
			}
		}
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

func sampleRows() [][]string {
	return [][]string{
		{"RUKO 2 LANTAI"},
		{},
		{"NO", "DIVISI", "URAIAN PEKERJAAN", "SATUAN", "VOLUME", "HARGA SATUAN", "JUMLAH"},
		{"1", "A", "Pembersihan lahan", "m2", "120", "15.000", "1.800.000"},
		{"2", "B", "Pondasi batu kali", "m3", "18,5", "850000", "15.725.000"},
		{"3", "E", "Atap spandek", "m2", "60.5", "120000", ""},
		{},
		{"", "", "SUB TOTAL", "", "", "", "17.525.000"},
		{"", "", "PPN 11%", "", "", "", "1.927.750"},
		{"", "", "TOTAL RAB", "", "", "", "19.452.750"},
		{},
		{"REKAP PER DIVISI"},
		{"DIVISI", "JUMLAH"},
		{"A", "1.800.000"},
		{"B", "15.725.000"},
		{"E", "7.260.000"},
		{"CATATAN", "Harga berlaku s/d akhir tahun"},
	}
}

func TestParseSampleDoc(t *testing.T) {
	doc, err := Parse(buildDoc(t, sampleRows()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc.Title != "RUKO 2 LANTAI" {
		t.Errorf("Title = %q", doc.Title)
	}
	if len(doc.Items) != 3 {
		t.Fatalf("items = %d want 3", len(doc.Items))
	}

	it := doc.Items[0]
	if it.Division != "A" || it.Description != "Pembersihan lahan" || it.Unit != "m2" {
		t.Errorf("item0 = %+v", it)
	}
	if !it.Volume.Equal(decimal.NewFromInt(120)) {
		t.Errorf("item0 volume = %s", it.Volume)
	}
	if !it.UnitPrice.Equal(decimal.NewFromInt(15000)) {
		t.Errorf("item0 unitPrice = %s", it.UnitPrice)
	}
	if !it.TotalCost.Equal(decimal.NewFromInt(1800000)) {
		t.Errorf("item0 totalCost = %s", it.TotalCost)
	}
	if it.Category != "preparation" {
		t.Errorf("item0 category = %s", it.Category)
	}

	// comma decimal volume "18,5" → 18.5
	if !doc.Items[1].Volume.Equal(decimal.RequireFromString("18.5")) {
		t.Errorf("item1 volume = %s want 18.5", doc.Items[1].Volume)
	}
	// thousands separator "15.000" → 15000
	if !doc.Items[1].UnitPrice.Equal(decimal.NewFromInt(850000)) {
		t.Errorf("item1 unitPrice = %s", doc.Items[1].UnitPrice)
	}
	// JUMLAH empty → vol × hs
	if !doc.Items[2].TotalCost.Equal(decimal.RequireFromString("7260000")) {
		t.Errorf("item2 totalCost computed = %s want 7.260.000", doc.Items[2].TotalCost)
	}

	if !doc.Subtotal.Equal(decimal.NewFromInt(17525000)) {
		t.Errorf("Subtotal = %s", doc.Subtotal)
	}
	if !doc.PPN.Equal(decimal.NewFromInt(1927750)) {
		t.Errorf("PPN = %s", doc.PPN)
	}
	if !doc.Total.Equal(decimal.NewFromInt(19452750)) {
		t.Errorf("Total = %s", doc.Total)
	}

	if len(doc.Divisions) != 3 {
		t.Fatalf("divisions = %d want 3", len(doc.Divisions))
	}
	if doc.Divisions[0].Letter != "A" || !doc.Divisions[0].Amount.Equal(decimal.NewFromInt(1800000)) {
		t.Errorf("division[0] = %+v", doc.Divisions[0])
	}
	if doc.Divisions[1].Name != "Struktur & Pondasi" {
		t.Errorf("division[1].Name = %q", doc.Divisions[1].Name)
	}
}

func TestParseUnknownDivisionLetter(t *testing.T) {
	rows := sampleRows()
	rows[3][1] = "Z"
	doc, err := Parse(buildDoc(t, rows))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(doc.Items) != 2 {
		t.Fatalf("items = %d want 2 (Z filtered)", len(doc.Items))
	}
}

func TestParsePicksSheetContainingBOQ(t *testing.T) {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Data Mentah")
	f.NewSheet("BOQ & RAB 2026")
	f.SetCellValue("BOQ & RAB 2026", "A1", "Proyek X")
	f.SetCellValue("BOQ & RAB 2026", "A3", "URAIAN PEKERJAAN SATUAN VOLUME HARGA SATUAN JUMLAH")
	f.SetCellValue("BOQ & RAB 2026", "A4", "Galian tanah")
	f.SetCellValue("BOQ & RAB 2026", "B4", "A")
	f.SetCellValue("BOQ & RAB 2026", "C4", "Galian tanah")
	f.SetCellValue("BOQ & RAB 2026", "D4", "m3")
	f.SetCellValue("BOQ & RAB 2026", "E4", "10")
	f.SetCellValue("BOQ & RAB 2026", "F4", "50000")
	t.Cleanup(func() { f.Close() })

	doc, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc.Title != "Proyek X" {
		t.Errorf("Title = %q", doc.Title)
	}
	if len(doc.Items) != 1 || doc.Items[0].Description != "Galian tanah" {
		t.Errorf("items = %+v", doc.Items)
	}
}

func TestParseMissingHeader(t *testing.T) {
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "no header here")
	t.Cleanup(func() { f.Close() })
	_, err := Parse(f)
	if err == nil {
		t.Fatal("expected error for missing header")
	}
}

func TestParseNoItems(t *testing.T) {
	rows := [][]string{
		{"X"},
		{},
		{"NO", "DIVISI", "URAIAN PEKERJAAN", "SATUAN", "VOLUME", "HARGA SATUAN", "JUMLAH"},
		{},
	}
	_, err := Parse(buildDoc(t, rows))
	if err == nil {
		t.Fatal("expected error when no items found")
	}
}

func TestParseUnitAliases(t *testing.T) {
	rows := sampleRows()
	rows[3][3] = "m'"
	rows[4][3] = "buah"
	rows[5][3] = "m3"
	doc, err := Parse(buildDoc(t, rows))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc.Items[0].Unit != "m1" {
		t.Errorf("m' → %q want m1", doc.Items[0].Unit)
	}
	if doc.Items[1].Unit != "unit" {
		t.Errorf("buah → %q want unit", doc.Items[1].Unit)
	}
	if doc.Items[2].Unit != "m3" {
		t.Errorf("m3 → %q", doc.Items[2].Unit)
	}
}

func TestParseEmptyUnitDefaults(t *testing.T) {
	rows := sampleRows()
	rows[3][3] = ""
	doc, err := Parse(buildDoc(t, rows))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc.Items[0].Unit != "unit" {
		t.Errorf("empty unit → %q want unit", doc.Items[0].Unit)
	}
}

func TestParseSkipsUnparsableRows(t *testing.T) {
	rows := sampleRows()
	rows[4][4] = "banana" // volume unparsable → row skipped
	doc, err := Parse(buildDoc(t, rows))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(doc.Items) != 2 {
		t.Fatalf("items = %d want 2", len(doc.Items))
	}
}

func TestParseRekapStopsAtCatatan(t *testing.T) {
	rows := sampleRows()
	rows = append(rows, [][]string{
		{"G", "5.000.000"}, // after CATATAN — must be ignored
	}...)
	doc, err := Parse(buildDoc(t, rows))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(doc.Divisions) != 3 {
		t.Fatalf("divisions = %d want 3 (post-CATATAN rows ignored)", len(doc.Divisions))
	}
}

func TestParseNum(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"60.5", "60.5", true},
		{"1.000.000", "1000000", true},
		{"18,5", "18.5", true},              // Indonesian comma decimal
		{"2,625", "2.625", true},            // decimal < 1
		{"1.800.000,50", "1800000.5", true}, // comma decimal + thousands dots
		{"Rp 15.000", "15000", true},        // single-dot thousands (price)
		{"15.000", "15000", true},
		{"1.500", "1500", true}, // 3-digit group → thousands
		{"15000", "15000", true},
		{"", "", false},
		{"abc", "", false},
		{"1.2.3", "123", true}, // multiple dots → thousands
	}
	for _, c := range cases {
		got, err := parseNum(c.in)
		if c.ok != (err == nil) {
			t.Errorf("parseNum(%q) err = %v, want ok=%v", c.in, err, c.ok)
			continue
		}
		if c.ok && !got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("parseNum(%q) = %s want %s", c.in, got, c.want)
		}
	}
}

func TestNormalizeUnit(t *testing.T) {
	cases := map[string]string{
		"M'":    "m1",
		"m2":    "m2",
		"M²":    "m2",
		"m³":    "m3",
		"BUAH":  "unit",
		"Bh":    "unit",
		"unit":  "unit",
		"ls":    "ls",
		"":      "unit",
		"  kg ": "kg",
	}
	for in, want := range cases {
		if got := normalizeUnit(in); got != want {
			t.Errorf("normalizeUnit(%q) = %q want %q", in, got, want)
		}
	}
}

func TestLastNum(t *testing.T) {
	n, err := lastNum([]string{"TOTAL RAB", "", "19.452.750"})
	if err != nil {
		t.Fatalf("lastNum: %v", err)
	}
	if !n.Equal(decimal.NewFromInt(19452750)) {
		t.Errorf("lastNum = %s", n)
	}
	if _, err := lastNum([]string{"", ""}); err == nil {
		t.Error("expected error for empty row")
	}
}

func TestPad(t *testing.T) {
	got := pad([]string{"a"}, 3)
	if len(got) != 3 || got[0] != "a" || got[1] != "" || got[2] != "" {
		t.Errorf("pad = %#v", got)
	}
	if p := pad([]string{"a", "b", "c"}, 2); len(p) != 3 || !strings.EqualFold(p[2], "c") {
		t.Errorf("pad should not truncate: %#v", p)
	}
}
