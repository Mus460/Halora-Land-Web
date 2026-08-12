package ahsp

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestNormalizeNum(t *testing.T) {
	cases := map[string]string{
		"100000":                "100000",
		"2,625":                 "2.625",
		"2,625.00":              "2625.00",
		"Rp 1.500.000,50":       "1.500.00050", // mixed thousands-dot + comma: comma treated as thousands separator
		"1.6899999999999998E-2": "1.6899999999999998E-2",
		" 5.000 ":               "5.000",
		"Rp":                    "",
		"12,5":                  "12.5",
		"18,5":                  "18.5",
	}
	for in, want := range cases {
		if got := normalizeNum(in); got != want {
			t.Errorf("normalizeNum(%q) = %q want %q", in, got, want)
		}
	}
	if _, err := decimal.NewFromString(normalizeNum("2,625")); err != nil {
		t.Errorf("normalizeNum(2,625) not parseable: %v", err)
	}
}

func TestNormalizeKode(t *testing.T) {
	cases := map[string]string{
		"":        "",
		"1.1.1.1": "1.1.1.1",
		"2.10":    "2.1",
		"2.10.0":  "2.10.0",
		"3.0":     "3",
		"1.1.0":   "1.1.0",
	}
	for in, want := range cases {
		if got := normalizeKode(in); got != want {
			t.Errorf("normalizeKode(%q) = %q want %q", in, got, want)
		}
	}
}

func TestExtractUnit(t *testing.T) {
	cases := map[string]string{
		"Pasir pasang 1 m3":    "m3",
		"Tukang batu 0,2 OH":   "OH",
		"Paku 5 kg":            "kg",
		"Semen portland 1 zak": "unit",
		"Tenaga kerja":         "unit",
		"Bata merah m'":        "m",
		"":                     "unit",
	}
	for in, want := range cases {
		if got := extractUnit(in); got != want {
			t.Errorf("extractUnit(%q) = %q want %q", in, got, want)
		}
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Rangka Atap":            "rangka-atap",
		"  Pekerjaan  Dinding  ": "pekerjaan-dinding",
		"Bata_Hitam!":            "bata-hitam",
		"":                       "",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q want %q", in, got, want)
		}
	}
}

func TestSectionType(t *testing.T) {
	if sectionType("A") != typeLabor {
		t.Error("A != labor")
	}
	if sectionType("B") != typeMaterial {
		t.Error("B != material")
	}
	if sectionType("C") != typeEquipment {
		t.Error("C != equipment")
	}
	if sectionType("D") != "" || sectionType("") != "" {
		t.Error("unknown marker must map to empty")
	}
}

func TestPadCols(t *testing.T) {
	if got := padCols([]string{"a", "b"}, 4); len(got) != 4 || got[0] != "a" || got[3] != "" {
		t.Errorf("padCols = %#v", got)
	}
	if got := padCols([]string{"a", "b", "c"}, 2); len(got) != 3 || got[2] != "c" {
		t.Errorf("padCols short = %#v", got)
	}
}

func TestDetectCodeCol(t *testing.T) {
	rows := [][]string{
		{"no", "kode", "kode ref", "uraian"},
		{"1", "2.1.1.1", "M.1", "Pasang"},
		{"2", "2.1.1.2", "M.2", "Pasang"},
	}
	if got := detectCodeCol(rows); got != 1 {
		t.Errorf("detectCodeCol = %d want 1", got)
	}
}

func TestDetectLayout(t *testing.T) {
	lay := detectLayout([]string{"No", "KODE", "KOEF", "SAT", "HARGA", "URAIAN"})
	if lay.description != 5 || lay.unit != 3 || lay.price != 4 || lay.coefficient != 2 {
		t.Errorf("lay = %+v", lay)
	}
	if lay.kodeRef != 1 {
		t.Errorf("kodeRef = %d", lay.kodeRef)
	}
	def := detectLayout(nil)
	if def.description != 3 || def.kodeRef != -1 || def.unit != 5 || def.coefficient != 6 || def.price != 7 {
		t.Errorf("default lay = %+v", def)
	}
}

func TestRegexPatterns(t *testing.T) {
	valid := []string{"2.1", "2.1.1.1", "10.20.30"}
	for _, v := range valid {
		if !kodeRe.MatchString(v) {
			t.Errorf("kodeRe(%q) should match", v)
		}
	}
	invalid := []string{"", "1", "a.1", "1.a", ".1", "1.", "1..2"}
	for _, v := range invalid {
		if kodeRe.MatchString(v) {
			t.Errorf("kodeRe(%q) should NOT match", v)
		}
	}
	if !rowNoRe.MatchString("42") || rowNoRe.MatchString("42a") {
		t.Error("rowNoRe broken")
	}
	for _, v := range []string{"I", "I.", "II", "II.", "III", "III."} {
		if !sectionRe.MatchString(v) {
			t.Errorf("sectionRe(%q) should match", v)
		}
	}
	for _, v := range []string{"IV", "", "X"} {
		if sectionRe.MatchString(v) {
			t.Errorf("sectionRe(%q) should NOT match", v)
		}
	}
}

func TestSheetToCategoryMapping(t *testing.T) {
	if SheetToCategory["Beton"] != "beton" {
		t.Error("Beton != beton")
	}
	if SheetToCategory["Rangka Atap"] != "rangka-atap" {
		t.Error("Rangka Atap != rangka-atap")
	}
	if len(SheetToCategory) == 0 {
		t.Error("empty mapping")
	}
}

func TestIsMason(t *testing.T) {
	if !isMason("Tukang batu") || !isMason("  TUKANG BATU  ") {
		t.Error("mason names should match")
	}
	if isMason("Tukang kayu") || isMason("") {
		t.Error("non-mason names must not match")
	}
}
