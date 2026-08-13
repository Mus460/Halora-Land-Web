package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

func analisaRow() []string {
	return []string{"id", "code", "name", "level", "parentId", "unit", "unitPrice", "category",
		"isGlobal", "userId", "isSystem", "ahspCode", "ahspSheet", "generalCost", "createdAt", "updatedAt"}
}

func componentRow() []string {
	return []string{"id", "analysisMasterId", "componentId", "coefficient", "type", "name", "unit",
		"unitPrice", "totalPrice", "referenceCode", "duration", "sequence", "createdAt", "updatedAt"}
}

func detailRow() []string {
	return []string{"id", "workItemId", "priceMasterId", "analysisMasterId", "name", "unit", "coefficient",
		"unitPrice", "totalCost", "type", "snapshotAt", "sourceCode"}
}

func workItemRowSvc() []string {
	return []string{"id", "projectId", "category", "description", "volume", "unit", "unitPrice", "totalCost",
		"calculationMethod", "level", "type", "analysisMasterId", "basePrice", "duration", "totalDuration", "createdAt", "updatedAt"}
}

func snapshotService(m pgxmock.PgxPoolIface) *SnapshotService {
	return NewSnapshotService(m, repository.NewWorkItemRepo(m), repository.NewAnalysisMasterRepo(m), nil)
}

func TestSnapshotFromAHSPHappyPath(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(9), "2.1.1.1", "Rangka Atap", 2, nil, "m2", "522451.64", "roof",
				true, nil, true, "2.1.1.1", "Atap", "0", time.Now(), time.Now()))
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows([]string{"s"}).AddRow("3.5"))
	m.ExpectBegin()
	lvl := "2"
	m.ExpectQuery(`INSERT INTO work_items`).
		WithArgs(int32(1), models.CategoryRoof, "Rangka Atap", "10", "m2", "522451.64", "5224516.4",
			models.MethodAHSP, &lvl, (*string)(nil), int32Ptr(9), "522451.64", "3.5").
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryRoof, "Rangka Atap", "10", "m2", "522451.64", "5224516.4",
				models.MethodAHSP, "2", nil, int32(9), nil, "3.5", "35", time.Now(), time.Now()))
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(componentRow()).
			AddRow(int32(1), int32(9), int32(20), "0.2", models.ComponentLabor, "Tukang", "OH", "145000", "29000", "L.02", "2", 1, time.Now(), time.Now()).
			AddRow(int32(2), int32(9), nil, "2.625", models.ComponentMaterial, "Pasir", "kg", "300", "787.5", nil, nil, 2, time.Now(), time.Now()))
	m.ExpectExec(`INSERT INTO work_item_details`).
		WithArgs(int32(7), int32Ptr(20), int32Ptr(9), "Tukang", "OH", "0.2", "145000", "290000", models.ComponentLabor, strp("L.02")).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectExec(`INSERT INTO work_item_details`).
		WithArgs(int32(7), (*int32)(nil), int32Ptr(9), "Pasir", "kg", "2.625", "300", "7875", models.ComponentMaterial, (*string)(nil)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectCommit()
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryRoof, "Rangka Atap", "10", "m2", "522451.64", "5224516.4",
				models.MethodAHSP, "2", nil, int32(9), nil, "3.5", "35", time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE "workItemId"`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()).
			AddRow(int32(1), int32(7), int32(20), int32(9), "Tukang", "OH", "0.2", "145000", "290000", models.ComponentLabor, time.Now(), "L.02"))

	s := snapshotService(m)
	it, err := s.FromAHSP(context.Background(), 1, 9, decimal.NewFromInt(10), true, nil)
	if err != nil {
		t.Fatalf("FromAHSP: %v", err)
	}
	if it.ID != 7 || it.Category != models.CategoryRoof {
		t.Errorf("it = %+v", it)
	}
	if !it.TotalCost.Equal(decimal.RequireFromString("5224516.4")) {
		t.Errorf("totalCost = %s", it.TotalCost)
	}
	if it.Duration == nil || !it.Duration.Equal(decimal.RequireFromString("3.5")) {
		t.Errorf("duration = %v", it.Duration)
	}
	if it.Level == nil || *it.Level != "2" {
		t.Errorf("level = %v", it.Level)
	}
	if len(it.ItemDetails) != 1 {
		t.Errorf("details = %d", len(it.ItemDetails))
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestSnapshotFromAHSPWithoutBreakdown(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(9), "1.1", "Pembersihan", 1, nil, "m'", "100000", "preparation",
				true, nil, true, "1.1", "Persiapan", "0", time.Now(), time.Now()))
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows([]string{"s"}).AddRow(nil))
	m.ExpectBegin()
	lvl := "1"
	m.ExpectQuery(`INSERT INTO work_items`).
		WithArgs(int32(1), models.CategoryPreparation, "Pembersihan", "2", "m'", "100000", "200000",
			models.MethodAHSP, &lvl, (*string)(nil), int32Ptr(9), "100000", nil).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(8), int32(1), models.CategoryPreparation, "Pembersihan", "2", "m'", "100000", "200000",
				models.MethodAHSP, "1", nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectCommit()
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(8)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(8), int32(1), models.CategoryPreparation, "Pembersihan", "2", "m'", "100000", "200000",
				models.MethodAHSP, "1", nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE "workItemId"`).WithArgs(int32(8)).
		WillReturnRows(pgxmock.NewRows(detailRow()))

	s := snapshotService(m)
	it, err := s.FromAHSP(context.Background(), 1, 9, decimal.NewFromInt(2), false, nil)
	if err != nil {
		t.Fatalf("FromAHSP: %v", err)
	}
	if len(it.ItemDetails) != 0 {
		t.Errorf("details = %d", len(it.ItemDetails))
	}
	if it.Level == nil || *it.Level != "1" {
		t.Errorf("level = %v", it.Level)
	}
}

func TestSnapshotFromAHSPNilMasterFields(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(5)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(5), "X.1", "Custom", 3, nil, nil, nil, nil, false, int32(2), false, nil, nil, "0", time.Now(), time.Now()))
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(5)).
		WillReturnRows(pgxmock.NewRows([]string{"s"}).AddRow("1"))
	m.ExpectBegin()
	lvl := "3"
	m.ExpectQuery(`INSERT INTO work_items`).
		WithArgs(int32(1), models.CategoryCustom, "Custom", "5", "unit", "0", "0",
			models.MethodAHSP, &lvl, (*string)(nil), int32Ptr(5), nil, "1").
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(5), int32(1), models.CategoryCustom, "Custom", "5", "unit", "0", "0",
				models.MethodAHSP, "3", nil, int32(5), nil, "1", "5", time.Now(), time.Now()))
	m.ExpectCommit()
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(5)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(5), int32(1), models.CategoryCustom, "Custom", "5", "unit", "0", "0",
				models.MethodAHSP, "3", nil, int32(5), nil, "1", "5", time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE "workItemId"`).WithArgs(int32(5)).
		WillReturnRows(pgxmock.NewRows(detailRow()))

	s := snapshotService(m)
	it, err := s.FromAHSP(context.Background(), 1, 5, decimal.NewFromInt(5), false, nil)
	if err != nil {
		t.Fatalf("FromAHSP: %v", err)
	}
	if it.Unit != "unit" || it.Category != models.CategoryCustom {
		t.Errorf("it = %+v", it)
	}
	if !it.UnitPrice.IsZero() {
		t.Errorf("unitPrice = %s", it.UnitPrice)
	}
}

func TestSnapshotFromAHSPErrors(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(m pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "master load error",
			setup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).WillReturnError(pgx.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "duration load error",
			setup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).
					WillReturnRows(pgxmock.NewRows(analisaRow()).
						AddRow(int32(9), "1.1", "Pembersihan", 1, nil, "m'", "100000", "preparation", true, nil, true, "1.1", "Persiapan", "0", time.Now(), time.Now()))
				m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).WillReturnError(context.Canceled)
			},
			wantErr: true,
		},
		{
			name: "create error",
			setup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).
					WillReturnRows(pgxmock.NewRows(analisaRow()).
						AddRow(int32(9), "1.1", "Pembersihan", 1, nil, "m'", "100000", "preparation", true, nil, true, "1.1", "Persiapan", "0", time.Now(), time.Now()))
				m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
					WillReturnRows(pgxmock.NewRows([]string{"s"}).AddRow(nil))
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO work_items`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(context.Canceled)
			},
			wantErr: true,
		},
		{
			name: "components load error",
			setup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).
					WillReturnRows(pgxmock.NewRows(analisaRow()).
						AddRow(int32(9), "1.1", "Pembersihan", 1, nil, "m'", "100000", "preparation", true, nil, true, "1.1", "Persiapan", "0", time.Now(), time.Now()))
				m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
					WillReturnRows(pgxmock.NewRows([]string{"s"}).AddRow(nil))
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO work_items`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
						AddRow(int32(7), int32(1), models.CategoryPreparation, "Pembersihan", "2", "m'", "100000", "200000", models.MethodAHSP, "1", nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
				m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).WillReturnError(context.Canceled)
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newPool(t)
			tc.setup(m)
			s := snapshotService(m)
			_, err := s.FromAHSP(context.Background(), 1, 9, decimal.NewFromInt(2), true, nil)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestSnapshotValidateNoDrift(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryConcrete, "Beton", "1", "m3", "928002", "928002", models.MethodAHSP, nil, nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()).
			AddRow(int32(1), int32(7), int32(20), int32(9), "Semen", "kg", "420", "1750", "735000", models.ComponentMaterial, time.Now(), "M.10"))
	m.ExpectQuery(`FROM price_masters WHERE id = ANY`).WithArgs([]int32{20}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "price"}).AddRow(int32(20), "1750"))

	s := snapshotService(m)
	res, err := s.ValidateSnapshot(context.Background(), 7)
	if err != nil {
		t.Fatalf("ValidateSnapshot: %v", err)
	}
	if !res.IsValid || len(res.Changes) != 0 {
		t.Errorf("res = %+v", res)
	}
}

func TestSnapshotValidateDetectsDrift(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryConcrete, "Beton", "1", "m3", "928002", "928002", models.MethodAHSP, nil, nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()).
			AddRow(int32(1), int32(7), int32(20), int32(9), "Semen", "kg", "420", "1000", "420000", models.ComponentMaterial, time.Now(), "M.10").
			AddRow(int32(2), int32(7), nil, nil, "Manual", "ls", "1", "500", "500", models.ComponentMaterial, time.Now(), nil).
			AddRow(int32(3), int32(7), int32(99), int32(9), "Hapus", "kg", "1", "100", "100", models.ComponentMaterial, time.Now(), nil))
	m.ExpectQuery(`FROM price_masters WHERE id = ANY`).WithArgs([]int32{20, 99}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "price"}).
			AddRow(int32(20), "1500").
			AddRow(int32(99), "250"))

	s := snapshotService(m)
	res, err := s.ValidateSnapshot(context.Background(), 7)
	if err != nil {
		t.Fatalf("ValidateSnapshot: %v", err)
	}
	if res.IsValid {
		t.Error("expected drift detected")
	}
	// nil priceMasterId skipped; id 99 present in live map with different price → counted
	if len(res.Changes) != 2 {
		t.Fatalf("changes = %+v", res.Changes)
	}
	c := res.Changes[0]
	if !c.Diff.Equal(decimal.NewFromInt(500)) {
		t.Errorf("diff = %s", c.Diff)
	}
	if !c.PercentChange.Equal(decimal.NewFromInt(50)) {
		t.Errorf("pct = %s", c.PercentChange)
	}
	if !c.OldPrice.Equal(decimal.NewFromInt(1000)) || !c.NewPrice.Equal(decimal.NewFromInt(1500)) {
		t.Errorf("prices = %s -> %s", c.OldPrice, c.NewPrice)
	}
}

func TestSnapshotValidateNoDetails(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategorySteel, "Manual", "1", "ls", "1000", "1000", models.MethodManual, nil, nil, nil, nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()))
	s := snapshotService(m)
	res, err := s.ValidateSnapshot(context.Background(), 7)
	if err != nil {
		t.Fatalf("ValidateSnapshot: %v", err)
	}
	if !res.IsValid {
		t.Errorf("res = %+v", res)
	}
}

func TestSnapshotValidateGetError(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(99)).WillReturnError(pgx.ErrNoRows)
	s := snapshotService(m)
	if _, err := s.ValidateSnapshot(context.Background(), 99); err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotRecalculate(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryConcrete, "Beton", "10", "m3", "800000", "8000000", models.MethodAHSP, nil, nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()).
			AddRow(int32(1), int32(7), int32(20), int32(9), "Semen", "kg", "420", "1750", "735000", models.ComponentMaterial, time.Now(), "M.10"))
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(9), "2.1.1.1", "Rangka Atap", 2, nil, "m2", "522451.64", "roof",
				true, nil, true, "2.1.1.1", "Atap", "0", time.Now(), time.Now()))
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(componentRow()).
			AddRow(int32(1), int32(9), int32(20), "0.2", models.ComponentLabor, "Tukang", "OH", "145000", "29000", "L.02", "2", 1, time.Now(), time.Now()).
			AddRow(int32(2), int32(9), int32(21), "2.625", models.ComponentMaterial, "Pasir", "kg", "300", "787.5", nil, nil, 2, time.Now(), time.Now()))
	m.ExpectQuery(`FROM price_masters WHERE id = ANY`).WithArgs([]int32{20, 21}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "price"}).
			AddRow(int32(20), "150000").
			AddRow(int32(21), "320"))
	m.ExpectBegin()
	m.ExpectExec(`DELETE FROM work_item_details`).WithArgs(int32(7)).
		WillReturnResult(pgxmock.NewResult("DELETE", 2))
	m.ExpectExec(`INSERT INTO work_item_details`).
		WithArgs(int32(7), int32Ptr(20), int32Ptr(9), "Tukang", "OH", "0.2", "150000", "300000", models.ComponentLabor, strp("L.02")).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectExec(`INSERT INTO work_item_details`).
		WithArgs(int32(7), int32Ptr(21), int32Ptr(9), "Pasir", "kg", "2.625", "320", "8400", models.ComponentMaterial, (*string)(nil)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectExec(`UPDATE work_items SET "unitPrice"`).
		WithArgs(int32(7), "30840", "308400").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	m.ExpectExec(`UPDATE work_items SET "basePrice"`).WithArgs(int32(7), "522451.64").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	m.ExpectCommit()
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryConcrete, "Beton", "10", "m3", "30840", "308400", models.MethodAHSP, nil, nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE "workItemId"`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()))

	s := snapshotService(m)
	it, err := s.Recalculate(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("Recalculate: %v", err)
	}
	if !it.UnitPrice.Equal(decimal.NewFromInt(30840)) {
		t.Errorf("unitPrice = %s", it.UnitPrice)
	}
	if !it.TotalCost.Equal(decimal.NewFromInt(308400)) {
		t.Errorf("totalCost = %s", it.TotalCost)
	}
}

func TestSnapshotRecalculateRejectsNonAHSP(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategorySteel, "Manual", "1", "ls", "1000", "1000", models.MethodManual, nil, nil, nil, nil, nil, nil, time.Now(), time.Now()))
	s := snapshotService(m)
	if _, err := s.Recalculate(context.Background(), 7, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotRecalculateNoLineage(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryConcrete, "Beton", "1", "m3", "928002", "928002", models.MethodAHSP, nil, nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()).
			AddRow(int32(1), int32(7), int32(20), nil, "Semen", "kg", "420", "1750", "735000", models.ComponentMaterial, time.Now(), "M.10"))
	s := snapshotService(m)
	if _, err := s.Recalculate(context.Background(), 7, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestSnapshotRecalculateAll(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryConcrete, "Beton", "10", "m3", "800000", "8000000", models.MethodAHSP, nil, nil, int32(9), nil, nil, nil, time.Now(), time.Now()).
			AddRow(int32(8), int32(1), models.CategorySteel, "Manual", "1", "ls", "1000", "1000", models.MethodManual, nil, nil, nil, nil, nil, nil, time.Now(), time.Now()))

	// Recalculate for item 7 only
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryConcrete, "Beton", "10", "m3", "800000", "8000000", models.MethodAHSP, nil, nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()).
			AddRow(int32(1), int32(7), int32(20), int32(9), "Semen", "kg", "420", "1750", "735000", models.ComponentMaterial, time.Now(), "M.10"))
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(9), "2.1.1.1", "Rangka Atap", 2, nil, "m2", "522451.64", "roof",
				true, nil, true, "2.1.1.1", "Atap", "0", time.Now(), time.Now()))
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(componentRow()).
			AddRow(int32(1), int32(9), int32(20), "1", models.ComponentMaterial, "Semen", "kg", "800", "800", "M.10", nil, 1, time.Now(), time.Now()))
	m.ExpectQuery(`FROM price_masters WHERE id = ANY`).WithArgs([]int32{20}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "price"}).AddRow(int32(20), "800"))
	m.ExpectBegin()
	m.ExpectExec(`DELETE FROM work_item_details`).WithArgs(int32(7)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	m.ExpectExec(`INSERT INTO work_item_details`).
		WithArgs(int32(7), int32Ptr(20), int32Ptr(9), "Semen", "kg", "1", "800", "8000", models.ComponentMaterial, strp("M.10")).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectExec(`UPDATE work_items SET "unitPrice"`).
		WithArgs(int32(7), "800", "8000").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	m.ExpectExec(`UPDATE work_items SET "basePrice"`).WithArgs(int32(7), "522451.64").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	m.ExpectCommit()
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(workItemRowSvc()).
			AddRow(int32(7), int32(1), models.CategoryConcrete, "Beton", "10", "m3", "800", "8000", models.MethodAHSP, nil, nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE "workItemId"`).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows(detailRow()))

	s := snapshotService(m)
	count, err := s.RecalculateAll(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("RecalculateAll: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d want 1", count)
	}
}

func int32Ptr(v int32) *int32 { return &v }

func strp(s string) *string { return &s }
