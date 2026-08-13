package service

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

func newPool(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	m, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	t.Cleanup(func() { m.Close() })
	return m
}

func workItemRow() []string {
	return []string{"id", "projectId", "category", "description", "volume", "unit", "unitPrice", "totalCost",
		"calculationMethod", "level", "type", "analysisMasterId", "basePrice", "duration", "totalDuration", "createdAt", "updatedAt"}
}

func TestRABComputeRollup(t *testing.T) {
	m := newPool(t)
	pid := int32(1)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows(workItemRow()).
			AddRow(int32(1), pid, models.CategoryFoundation, "Pondasi", "10", "m3", "1000000", "10000000", models.MethodAHSP, "2", nil, nil, nil, nil, nil, time.Now(), time.Now()).
			AddRow(int32(2), pid, models.CategoryRoof, "Atap", "20", "m2", "500000", "10000000", models.MethodManual, "3", nil, nil, nil, nil, nil, time.Now(), time.Now()).
			AddRow(int32(3), pid, models.CategoryRoof, "Rangka", "5", "m2", "2000000", "10000000", models.MethodAHSP, "3", nil, nil, nil, "2", "10", time.Now(), time.Now()))
	m.ExpectQuery(`SELECT margin::text FROM recaps`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows([]string{"margin"}).AddRow("15"))

	s := NewRABService(m, repository.NewWorkItemRepo(m), repository.NewRecapRepo(m),
		decimal.RequireFromString("0.10"), decimal.RequireFromString("0.11"))

	res, err := s.Compute(context.Background(), pid, &repository.SummaryProject{ID: pid, Name: "Ruko"})
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	// subtotal = 10M + 10M + 10M
	if !res.Summary.Subtotal.Equal(decimal.NewFromInt(30000000)) {
		t.Errorf("subtotal = %s", res.Summary.Subtotal)
	}
	// subtotalWithMargin = 30M × 1.15 = 34.5M
	if !res.Summary.SubtotalWithMargin.Equal(decimal.RequireFromString("34500000")) {
		t.Errorf("subtotalWithMargin = %s", res.Summary.SubtotalWithMargin)
	}
	// overhead = 34.5M × 0.10 = 3.45M
	if !res.Summary.Overhead.Equal(decimal.RequireFromString("3450000")) {
		t.Errorf("overhead = %s", res.Summary.Overhead)
	}
	// subtotalBeforeTax = 34.5M + 3.45M = 37.95M
	if !res.Summary.SubtotalBeforeTax.Equal(decimal.RequireFromString("37950000")) {
		t.Errorf("subtotalBeforeTax = %s", res.Summary.SubtotalBeforeTax)
	}
	// ppn = 37.95M × 0.11 = 4.1745M
	if !res.Summary.TotalPPN.Equal(decimal.RequireFromString("4174500")) {
		t.Errorf("totalPPN = %s", res.Summary.TotalPPN)
	}
	// totalFinal = 37.95M + 4.1745M
	if !res.Summary.TotalAkhir.Equal(decimal.RequireFromString("42124500")) {
		t.Errorf("totalFinal = %s", res.Summary.TotalAkhir)
	}
	if !res.Summary.Margin.Equal(decimal.RequireFromString("15")) {
		t.Errorf("margin = %s", res.Summary.Margin)
	}
	if !res.Summary.PPNPct.Equal(decimal.RequireFromString("11")) {
		t.Errorf("ppnPct = %s", res.Summary.PPNPct)
	}
	if !res.Summary.TotalDuration.Equal(decimal.NewFromInt(10)) {
		t.Errorf("totalDuration = %s", res.Summary.TotalDuration)
	}
	// grouping
	if len(res.Grouped) != 2 {
		t.Fatalf("grouped = %v", res.Grouped)
	}
	if len(res.Grouped["roof"]) != 2 || len(res.Grouped["foundation"]) != 1 {
		t.Errorf("grouped roof/foundation = %d/%d", len(res.Grouped["roof"]), len(res.Grouped["foundation"]))
	}
	if !res.Subtotals["roof"].Equal(decimal.NewFromInt(20000000)) {
		t.Errorf("roof subtotal = %s", res.Subtotals["roof"])
	}
	if !res.SubtotalDuration["roof"].Equal(decimal.NewFromInt(10)) {
		t.Errorf("roof duration = %s", res.SubtotalDuration["roof"])
	}
	// project summary
	if res.Project.Name != "Ruko" {
		t.Errorf("project = %+v", res.Project)
	}
}

func TestRABComputeZeroItems(t *testing.T) {
	m := newPool(t)
	pid := int32(1)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows(workItemRow()))
	m.ExpectQuery(`SELECT margin::text FROM recaps`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows([]string{"margin"}).AddRow(nil))

	s := NewRABService(m, repository.NewWorkItemRepo(m), repository.NewRecapRepo(m),
		decimal.RequireFromString("0.10"), decimal.RequireFromString("0.11"))
	res, err := s.Compute(context.Background(), pid, nil)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if !res.Summary.TotalAkhir.IsZero() {
		t.Errorf("totalFinal = %s want 0", res.Summary.TotalAkhir)
	}
	if res.Project.Name != "" {
		t.Errorf("project = %+v", res.Project)
	}
}

func TestRABComputePropagatesListError(t *testing.T) {
	m := newPool(t)
	pid := int32(1)
	m.ExpectQuery(`FROM work_items WHERE`).WithArgs(pid).WillReturnError(context.Canceled)
	s := NewRABService(m, repository.NewWorkItemRepo(m), repository.NewRecapRepo(m),
		decimal.RequireFromString("0.10"), decimal.RequireFromString("0.11"))
	if _, err := s.Compute(context.Background(), pid, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestRABComputeNegativeMargin(t *testing.T) {
	m := newPool(t)
	pid := int32(1)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows(workItemRow()).
			AddRow(int32(1), pid, models.CategoryFoundation, "Pondasi", "1", "m3", "1000000", "1000000", models.MethodAHSP, nil, nil, nil, nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`SELECT margin::text FROM recaps`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows([]string{"margin"}).AddRow("-5"))

	s := NewRABService(m, repository.NewWorkItemRepo(m), repository.NewRecapRepo(m),
		decimal.RequireFromString("0.10"), decimal.RequireFromString("0.11"))
	res, err := s.Compute(context.Background(), pid, nil)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	// 1M × 0.95 = 950k; overhead 95k; tax 11% of 1.045M
	if !res.Summary.SubtotalWithMargin.Equal(decimal.NewFromInt(950000)) {
		t.Errorf("subtotalWithMargin = %s", res.Summary.SubtotalWithMargin)
	}
}

func TestSortedCategories(t *testing.T) {
	got := SortedCategoryes(map[string]decimal.Decimal{
		"roof": decimal.Zero, "foundation": decimal.Zero, "custom": decimal.Zero,
	})
	if len(got) != 3 || got[0] != "custom" || got[1] != "foundation" || got[2] != "roof" {
		t.Errorf("got = %v", got)
	}
	if SortedCategoryes(nil) == nil {
		t.Error("nil map should produce empty (non-nil) slice")
	}
	if got := SortedCategoryes(map[string]decimal.Decimal{}); len(got) != 0 {
		t.Errorf("empty map = %v, want empty slice", got)
	}
}

func rabComputeWith(t *testing.T, margin string, overhead, ppn decimal.Decimal) *RecapResult {
	t.Helper()
	m := newPool(t)
	pid := int32(1)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows(workItemRow()).
			AddRow(int32(1), pid, models.CategoryFoundation, "Pondasi", "1", "m3", "1000000", "1000000", models.MethodAHSP, nil, nil, nil, nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`SELECT margin::text FROM recaps`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows([]string{"margin"}).AddRow(margin))
	s := NewRABService(m, repository.NewWorkItemRepo(m), repository.NewRecapRepo(m), overhead, ppn)
	res, err := s.Compute(context.Background(), pid, nil)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	return res
}

func TestRABComputeZeroRates(t *testing.T) {
	res := rabComputeWith(t, "0", decimal.Zero, decimal.Zero)
	if !res.Summary.Subtotal.Equal(decimal.NewFromInt(1000000)) {
		t.Errorf("subtotal = %s, want 1000000", res.Summary.Subtotal)
	}
	if !res.Summary.SubtotalWithMargin.Equal(decimal.NewFromInt(1000000)) {
		t.Errorf("subtotalWithMargin = %s, want 1000000", res.Summary.SubtotalWithMargin)
	}
	if !res.Summary.Overhead.IsZero() || !res.Summary.TotalPPN.IsZero() {
		t.Errorf("overhead/ppn = %s/%s, want 0/0", res.Summary.Overhead, res.Summary.TotalPPN)
	}
	if !res.Summary.TotalAkhir.Equal(decimal.NewFromInt(1000000)) {
		t.Errorf("totalFinal = %s, want 1000000", res.Summary.TotalAkhir)
	}
	if !res.Summary.PPNPct.IsZero() {
		t.Errorf("ppnPct = %s, want 0", res.Summary.PPNPct)
	}
}

func TestRABComputeFractionalMargin(t *testing.T) {
	res := rabComputeWith(t, "12.5", decimal.RequireFromString("0.10"), decimal.RequireFromString("0.11"))
	// 1M × 1.125 = 1.125M; overhead 112500; before tax 1.2375M; ppn 136125
	if !res.Summary.SubtotalWithMargin.Equal(decimal.RequireFromString("1125000")) {
		t.Errorf("subtotalWithMargin = %s", res.Summary.SubtotalWithMargin)
	}
	if !res.Summary.TotalPPN.Equal(decimal.RequireFromString("136125")) {
		t.Errorf("totalPPN = %s", res.Summary.TotalPPN)
	}
}

func TestRABComputeMarginMinusHundredWipesOut(t *testing.T) {
	res := rabComputeWith(t, "-100", decimal.RequireFromString("0.10"), decimal.RequireFromString("0.11"))
	if !res.Summary.SubtotalWithMargin.IsZero() {
		t.Errorf("subtotalWithMargin = %s, want 0", res.Summary.SubtotalWithMargin)
	}
	if !res.Summary.TotalAkhir.IsZero() {
		t.Errorf("totalFinal = %s, want 0", res.Summary.TotalAkhir)
	}
}

func TestRABComputeMarginBelowMinusHundredGoesNegative(t *testing.T) {
	// Extreme input: the math is not clamped, so the summary turns negative.
	// Documented behaviour — validation at the API layer is the guard.
	res := rabComputeWith(t, "-150", decimal.RequireFromString("0.10"), decimal.RequireFromString("0.11"))
	if !res.Summary.SubtotalWithMargin.Equal(decimal.RequireFromString("-500000")) {
		t.Errorf("subtotalWithMargin = %s, want -500000", res.Summary.SubtotalWithMargin)
	}
	if !res.Summary.Overhead.Equal(decimal.RequireFromString("-50000")) {
		t.Errorf("overhead = %s, want -50000", res.Summary.Overhead)
	}
}
