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

	s := NewRABService(m, repository.NewWorkItemRepo(m), decimal.RequireFromString("0.11"))

	res, err := s.Compute(context.Background(), pid, &repository.SummaryProject{ID: pid, Name: "Ruko"})
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	// subtotal = 10M + 10M + 10M
	if !res.Summary.Subtotal.Equal(decimal.NewFromInt(30000000)) {
		t.Errorf("subtotal = %s", res.Summary.Subtotal)
	}
	// ppn = 30M × 0.11 = 3.3M
	if !res.Summary.TotalPPN.Equal(decimal.RequireFromString("3300000")) {
		t.Errorf("totalPPN = %s", res.Summary.TotalPPN)
	}
	// totalFinal = 30M + 3.3M = 33.3M
	if !res.Summary.TotalAkhir.Equal(decimal.RequireFromString("33300000")) {
		t.Errorf("totalFinal = %s", res.Summary.TotalAkhir)
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

	s := NewRABService(m, repository.NewWorkItemRepo(m), decimal.RequireFromString("0.11"))
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
	s := NewRABService(m, repository.NewWorkItemRepo(m), decimal.RequireFromString("0.11"))
	if _, err := s.Compute(context.Background(), pid, nil); err == nil {
		t.Fatal("expected error")
	}
}

func rabComputeWith(t *testing.T, ppn decimal.Decimal, rows ...interface{}) *RecapResult {
	t.Helper()
	m := newPool(t)
	pid := int32(1)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows(workItemRow()).AddRow(rows...))
	s := NewRABService(m, repository.NewWorkItemRepo(m), ppn)
	res, err := s.Compute(context.Background(), pid, nil)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	return res
}

func TestRABComputeZeroRates(t *testing.T) {
	res := rabComputeWith(t, decimal.Zero,
		int32(1), int32(1), models.CategoryFoundation, "Pondasi", "1", "m3", "1000000", "1000000",
		models.MethodAHSP, nil, nil, nil, nil, nil, nil, time.Now(), time.Now())
	if !res.Summary.Subtotal.Equal(decimal.NewFromInt(1000000)) {
		t.Errorf("subtotal = %s, want 1000000", res.Summary.Subtotal)
	}
	if !res.Summary.TotalPPN.IsZero() {
		t.Errorf("totalPPN = %s, want 0", res.Summary.TotalPPN)
	}
	if !res.Summary.TotalAkhir.Equal(decimal.NewFromInt(1000000)) {
		t.Errorf("totalFinal = %s, want 1000000", res.Summary.TotalAkhir)
	}
	if !res.Summary.PPNPct.IsZero() {
		t.Errorf("ppnPct = %s, want 0", res.Summary.PPNPct)
	}
}

func TestRABComputeFractionalPPN(t *testing.T) {
	res := rabComputeWith(t, decimal.RequireFromString("0.11"),
		int32(1), int32(1), models.CategoryFoundation, "Pondasi", "1", "m3", "1000000", "1000000",
		models.MethodAHSP, nil, nil, nil, nil, nil, nil, time.Now(), time.Now())
	// 1M × 0.11 = 110k; totalFinal 1.11M
	if !res.Summary.TotalPPN.Equal(decimal.RequireFromString("110000")) {
		t.Errorf("totalPPN = %s", res.Summary.TotalPPN)
	}
	if !res.Summary.TotalAkhir.Equal(decimal.RequireFromString("1110000")) {
		t.Errorf("totalFinal = %s", res.Summary.TotalAkhir)
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

func TestRABSyncContractValue(t *testing.T) {
	m := newPool(t)
	pid := int32(1)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows(workItemRow()).
			AddRow(int32(1), pid, models.CategoryFoundation, "Pondasi", "1", "m3", "1000000", "1000000",
				models.MethodAHSP, nil, nil, nil, nil, nil, nil, time.Now(), time.Now()))
	m.ExpectExec(`UPDATE projects SET "contractValue"`).WithArgs(pid, "1110000").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	s := NewRABService(m, repository.NewWorkItemRepo(m), decimal.RequireFromString("0.11"))
	if err := s.SyncContractValue(context.Background(), pid); err != nil {
		t.Fatalf("SyncContractValue: %v", err)
	}
}