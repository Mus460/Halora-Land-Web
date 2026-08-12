package ahsp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func TestVerifyReportedBugs(t *testing.T) {
	f, err := excelize.OpenFile("../../data/ahsp-2026.xlsx")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	// 1. Rangka Atap 2.1.1.1 — material rows without row numbers must be
	// parsed (previously skipped by the row-number gate).
	items, _ := ParseSheet(f, "Rangka Atap")
	var ra *WorkItem
	for i := range items {
		if items[i].Code == "2.1.1.1" {
			ra = &items[i]
			break
		}
	}
	if ra == nil {
		t.Fatal("2.1.1.1 not found")
	}
	if len(ra.Breakdown) != 11 {
		t.Fatalf("2.1.1.1 breakdown = %d want 11 (labor+material incl. unnumbered rows)", len(ra.Breakdown))
	}
	mats := 0
	for _, b := range ra.Breakdown {
		if b.Type == "material" {
			mats++
			if b.UnitPrice.IsZero() {
				t.Errorf("material %q inline price missing", b.Name)
			}
		}
	}
	if mats != 7 {
		t.Errorf("2.1.1.1 materials = %d want 7 (baja ringan x3, dynabolt, screw x2, reng)", mats)
	}

	// 2. Beton — "Batu split 2/3" kg must use its inline unit price.
	bt, _ := ParseSheet(f, "Beton")
	checked := 0
	for _, it := range bt {
		for _, row := range it.Breakdown {
			if strings.Contains(row.Name, "Batu split 2/3") {
				if !row.UnitPrice.Equal(decimal.RequireFromString("260.96296296296299")) {
					t.Errorf("%s: %s %s inline price = %s want 260.96", it.Code, row.Name, row.Unit, row.UnitPrice)
				}
				if row.Type != "material" {
					t.Errorf("%s: type = %s want material", it.Code, row.Type)
				}
				checked++
				break
			}
		}
		if checked >= 2 {
			break
		}
	}
	if checked == 0 {
		t.Error("Batu split 2/3 not found in Beton")
	}

	// 3. Text decimal "2,625" (decimal comma in TEXT cells) parses as 2.625.
	tl, _ := ParseSheet(f, "Penutup Lantai dan Dinding")
	plint := 0
	for _, it := range tl {
		for _, row := range it.Breakdown {
			if row.Coefficient.String() == "2.625" {
				plint++
			}
		}
	}
	if plint < 8 {
		t.Errorf("items with koef 2.625 = %d want >= 8", plint)
	}
}

func TestVerifyUnitPriceComputation(t *testing.T) {
	f, err := excelize.OpenFile("../../data/ahsp-2026.xlsx")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	items, _ := ParseSheet(f, "Rangka Atap")
	for _, it := range items {
		if it.Code == "2.1.1.1" {
			total := decimal.Zero
			for _, b := range it.Breakdown {
				total = total.Add(b.Coefficient.Mul(b.UnitPrice))
			}
			hs := total.Mul(decimal.NewFromInt(1).Add(it.GeneralCost)).Round(2)
			fmt.Printf("2.1.1.1 sum=%s x1.10=%s (sheet D=474,956.04 F=522,451.64)\n", total, hs)
			if hs.LessThan(decimal.NewFromInt(522000)) || hs.GreaterThan(decimal.NewFromInt(523000)) {
				t.Errorf("unit price %s out of expected range", hs)
			}
		}
	}
}
