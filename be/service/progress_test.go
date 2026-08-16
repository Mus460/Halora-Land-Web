package service

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"
)

const progressQuery = `SELECT id, category, description, trim_scale\(ROUND\(volume, 2\)\)::text, unit, "totalCost"::text, progress, COALESCE\(duration, 0\)::text`

func progressRows() []string {
	return []string{"id", "category", "description", "volume", "unit", "totalCost", "progress", "duration"}
}

func expectProgressItems(m pgxmock.PgxPoolIface, projectID int32, rows [][]any) *pgxmock.Rows {
	cols := progressRows()
	rr := pgxmock.NewRows(cols)
	for _, r := range rows {
		rr.AddRow(r...)
	}
	m.ExpectQuery(progressQuery).WithArgs(projectID).WillReturnRows(rr)
	return rr
}

func TestProgressItemsWeightsRoundToTwoDecimals(t *testing.T) {
	m := newPool(t)
	expectProgressItems(m, int32(1), [][]any{
		{int32(1), "preparation", "A", "1", "bh", "100", 0, "8"},
		{int32(2), "preparation", "B", "1", "bh", "100", 0, "8"},
		{int32(3), "preparation", "C", "1", "bh", "100", 0, "8"},
	})
	items, total, err := NewProgressService(m).Items(context.Background(), 1)
	if err != nil {
		t.Fatalf("Items: %v", err)
	}
	if !total.Equal(decimal.NewFromInt(300)) {
		t.Errorf("total = %s, want 300", total.String())
	}
	if len(items) != 3 {
		t.Fatalf("len = %d, want 3", len(items))
	}
	for i, it := range items {
		if math.Abs(it.Weight-33.33) > 0.0001 {
			t.Errorf("item %d weight = %v, want 33.33", i, it.Weight)
		}
	}
}

func TestProgressItemsWeightUnbalanced(t *testing.T) {
	m := newPool(t)
	expectProgressItems(m, int32(1), [][]any{
		{int32(1), "foundation", "A", "1", "m3", "25", 50, "8"},
		{int32(2), "foundation", "B", "1", "m3", "75", 100, "8"},
	})
	items, _, err := NewProgressService(m).Items(context.Background(), 1)
	if err != nil {
		t.Fatalf("Items: %v", err)
	}
	if items[0].Weight != 25 || items[1].Weight != 75 {
		t.Errorf("weights = %v, %v; want 25, 75", items[0].Weight, items[1].Weight)
	}
	if items[0].Progress != 50 || items[1].Progress != 100 {
		t.Errorf("progress = %d, %d; want 50, 100", items[0].Progress, items[1].Progress)
	}
}

func TestProgressItemsZeroTotalLeavesWeightsZero(t *testing.T) {
	m := newPool(t)
	expectProgressItems(m, int32(1), [][]any{
		{int32(1), "preparation", "A", "1", "bh", "0", 0, "8"},
		{int32(2), "preparation", "B", "1", "bh", "0", 0, "8"},
	})
	items, total, err := NewProgressService(m).Items(context.Background(), 1)
	if err != nil {
		t.Fatalf("Items: %v", err)
	}
	if !total.IsZero() {
		t.Errorf("total = %s, want 0", total.String())
	}
	for i, it := range items {
		if it.Weight != 0 {
			t.Errorf("item %d weight = %v, want 0", i, it.Weight)
		}
	}
}

func TestProgressItemsHoursVolumeTimesDuration(t *testing.T) {
	m := newPool(t)
	expectProgressItems(m, int32(1), [][]any{
		{int32(1), "preparation", "A", "10", "bh", "100", 0, "5"},
		{int32(2), "preparation", "B", "0", "bh", "100", 0, "5"},
		{int32(3), "preparation", "C", "10", "bh", "100", 0, "0"},
		{int32(4), "preparation", "D", "10", "bh", "100", 0, ""},
		{int32(5), "preparation", "E", "", "bh", "100", 0, "5"},
		{int32(6), "preparation", "F", "2.5", "m3", "100", 0, "7.5"},
		{int32(7), "preparation", "G", "-1", "bh", "100", 0, "5"},
	})
	items, _, err := NewProgressService(m).Items(context.Background(), 1)
	if err != nil {
		t.Fatalf("Items: %v", err)
	}
	cases := []struct {
		idx  int
		want float64
	}{
		{0, 50},   // 10 × 5
		{1, 0},    // zero volume
		{2, 0},    // zero duration
		{3, 0},    // empty duration
		{4, 0},    // empty volume
		{5, 18.75}, // 2.5 × 7.5
	}
	for _, c := range cases {
		if math.Abs(items[c.idx].Hours-c.want) > 1e-9 {
			t.Errorf("item %d hours = %v, want %v", c.idx, items[c.idx].Hours, c.want)
		}
	}
}

func TestProgressItemsNegativeVolumeNoHours(t *testing.T) {
	m := newPool(t)
	expectProgressItems(m, int32(1), [][]any{
		{int32(1), "preparation", "G", "-1", "bh", "100", 0, "5"},
	})
	items, _, err := NewProgressService(m).Items(context.Background(), 1)
	if err != nil {
		t.Fatalf("Items: %v", err)
	}
	if items[0].Hours != 0 {
		t.Errorf("hours = %v, want 0", items[0].Hours)
	}
	if items[0].Weight != 100 {
		t.Errorf("weight = %v, want 100", items[0].Weight)
	}
}

func TestProgressItemsSkipsInvalidCostRows(t *testing.T) {
	m := newPool(t)
	expectProgressItems(m, int32(1), [][]any{
		{int32(1), "preparation", "A", "1", "bh", "abc", 0, "8"},
		{int32(2), "preparation", "B", "1", "bh", "50", 0, "8"},
	})
	items, total, err := NewProgressService(m).Items(context.Background(), 1)
	if err != nil {
		t.Fatalf("Items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1 (invalid cost row skipped)", len(items))
	}
	if items[0].ID != 2 || !total.Equal(decimal.NewFromInt(50)) {
		t.Errorf("items[0] = %+v, total = %s", items[0], total.String())
	}
	if items[0].Weight != 100 {
		t.Errorf("weight = %v, want 100", items[0].Weight)
	}
}

func TestProgressItemsEmptyProject(t *testing.T) {
	m := newPool(t)
	expectProgressItems(m, int32(99), [][]any{})
	items, total, err := NewProgressService(m).Items(context.Background(), 99)
	if err != nil {
		t.Fatalf("Items: %v", err)
	}
	if len(items) != 0 || !total.IsZero() {
		t.Errorf("items = %v, total = %s; want empty, 0", items, total.String())
	}
}

func TestProgressItemsQueryErrorPropagates(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(progressQuery).WithArgs(int32(1)).WillReturnError(errors.New("boom"))
	_, _, err := NewProgressService(m).Items(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProgressItemsScanErrorPropagates(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(progressQuery).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(progressRows()).
			AddRow("not-a-number", "preparation", "A", "1", "bh", "100", 0, "8"))
	_, _, err := NewProgressService(m).Items(context.Background(), 1)
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
}
