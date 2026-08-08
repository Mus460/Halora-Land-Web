package ahsp

import (
	"strings"

	"github.com/shopspring/decimal"
	"testing"

	"github.com/xuri/excelize/v2"
)

func openTestWorkbook(t *testing.T) *excelize.File {
	t.Helper()
	f, err := excelize.OpenFile("../../data/ahsp-2026.xlsx")
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func TestParsePriceList(t *testing.T) {
	f := openTestWorkbook(t)
	prices, err := ParsePriceList(f)
	if err != nil {
		t.Fatalf("ParsePriceList: %v", err)
	}
	if len(prices) < 2000 {
		t.Fatalf("expected >2000 price items, got %d", len(prices))
	}
	found := map[string][]PriceItem{}
	for _, p := range prices {
		key := strings.ToLower(strings.TrimSpace(p.Name))
		found[key] = append(found[key], p)
	}
	assert := func(name, sat string, price int64) {
		t.Helper()
		for _, p := range found[strings.ToLower(name)] {
			if p.Unit == sat {
				if !p.Price.Equal(decimal.NewFromInt(price)) {
					t.Errorf("price %q %s = %s want %d", name, sat, p.Price, price)
				}
				return
			}
		}
		t.Errorf("price %q unit %q not found", name, sat)
	}
	assert("Pekerja", "OH", 100000)
	assert("Tukang batu", "OH", 145000)
	assert("Pasir Beton", "kg", 300)
	assert("Pasir Beton", "m3", 370200)
	assert("Kerikil", "kg", 210)
}
func TestParseSheetPersiapan(t *testing.T) {
	f := openTestWorkbook(t)
	items, err := ParseSheet(f, "Persiapan")
	if err != nil {
		t.Fatalf("ParseSheet: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("no items parsed")
	}
	var found *WorkItem
	for i := range items {
		if items[i].Code == "1.1.1.1" {
			found = &items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("item 1.1.1.1 not parsed; got %d items", len(items))
	}
	if found.Unit != "m'" {
		t.Errorf("unit = %q want m'", found.Unit)
	}
	if len(found.Breakdown) < 10 {
		t.Fatalf("breakdown too small: %d", len(found.Breakdown))
	}
	byName := map[string]BreakdownRow{}
	for _, b := range found.Breakdown {
		byName[strings.ToLower(strings.TrimSpace(b.Name))] = b
	}
	tb, ok := byName["tukang batu"]
	if !ok {
		t.Fatal("Tukang batu row missing")
	}
	if !tb.Coefficient.Equal(decimal.NewFromFloat(0.2)) {
		t.Errorf("Tukang batu coefficient = %s want 0.2", tb.Coefficient)
	}
	if tb.ReferenceCode != "L.02" {
		t.Errorf("Tukang batu referenceCode = %q want L.02", tb.ReferenceCode)
	}
}

func TestParseSheetRisha(t *testing.T) {
	f := openTestWorkbook(t)
	items, err := ParseSheet(f, "Strutur Risha ")
	if err != nil {
		t.Fatalf("ParseSheet: %v", err)
	}
	if len(items) != 13 {
		t.Errorf("RISHA items = %d want 13", len(items))
	}
}
