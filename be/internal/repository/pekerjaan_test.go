package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
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

func TestWorkItemListWithProjectFilter(t *testing.T) {
	m := newPool(t)
	pid := int32(7)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).
		WithArgs(pid).
		WillReturnRows(pgxmock.NewRows(workItemRow()).
			AddRow(int32(1), pid, models.CategoryFoundation, "Pondasi", "12.5", "m3", "850000", "10625000",
				models.MethodAHSP, "2.1.1", nil, int32(3), nil, "4", "50", time.Now(), time.Now()))

	r := NewWorkItemRepo(m)
	items, err := r.List(context.Background(), ListWorkItemFilter{ProjectID: &pid})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d", len(items))
	}
	it := items[0]
	if !it.Volume.Equal(decimal.RequireFromString("12.5")) {
		t.Errorf("volume = %s", it.Volume)
	}
	if it.Category != models.CategoryFoundation {
		t.Errorf("category = %s", it.Category)
	}
	if it.Level == nil || *it.Level != "2.1.1" {
		t.Errorf("level = %v", it.Level)
	}
	if it.AnalysisMasterID == nil || *it.AnalysisMasterID != 3 {
		t.Errorf("analysisMasterId = %v", it.AnalysisMasterID)
	}
	if it.Duration == nil || !it.Duration.Equal(decimal.NewFromInt(4)) {
		t.Errorf("duration = %v", it.Duration)
	}
	if it.TotalDuration == nil || !it.TotalDuration.Equal(decimal.NewFromInt(50)) {
		t.Errorf("totalDuration = %v", it.TotalDuration)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestWorkItemListWithCategoryAndSearch(t *testing.T) {
	m := newPool(t)
	pid := int32(1)
	m.ExpectQuery(`FROM work_items WHERE "deletedAt" IS NULL AND "projectId"`).
		WithArgs(pid, "wall", "%cor%").
		WillReturnRows(pgxmock.NewRows(workItemRow()))
	r := NewWorkItemRepo(m)
	items, err := r.List(context.Background(), ListWorkItemFilter{ProjectID: &pid, Category: "wall", Search: "cor"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("items = %d want 0", len(items))
	}
}

func TestWorkItemGetIncludesDetails(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(5)).
		WillReturnRows(pgxmock.NewRows(workItemRow()).
			AddRow(int32(5), int32(1), models.CategoryConcrete, "Beton", "1", "m3", "928002", "928002",
				models.MethodAHSP, "2.2.1.4.1", nil, int32(9), nil, nil, nil, time.Now(), time.Now()))
	m.ExpectQuery(`FROM work_item_details WHERE "workItemId"`).WithArgs(int32(5)).
		WillReturnRows(pgxmock.NewRows(detailRow()).
			AddRow(int32(11), int32(5), int32(20), int32(9), "Semen", "kg", "420", "1750", "735000", models.ComponentMaterial, time.Now(), "M.10"))
	r := NewWorkItemRepo(m)
	it, err := r.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(it.ItemDetails) != 1 {
		t.Fatalf("details = %d", len(it.ItemDetails))
	}
	if it.ItemDetails[0].Coefficient.Equal(decimal.Zero) {
		t.Error("coefficient missing")
	}
	if it.ItemDetails[0].SourceCode == nil || *it.ItemDetails[0].SourceCode != "M.10" {
		t.Errorf("sourceCode = %v", it.ItemDetails[0].SourceCode)
	}
}

func detailRow() []string {
	return []string{"id", "workItemId", "priceMasterId", "analysisMasterId", "name", "unit", "coefficient",
		"unitPrice", "totalCost", "type", "snapshotAt", "sourceCode"}
}

func TestWorkItemGetByIDNoRows(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_items WHERE id =`).WithArgs(int32(99)).WillReturnError(pgx.ErrNoRows)
	r := NewWorkItemRepo(m)
	if _, err := r.GetByID(context.Background(), 99); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("err = %v want ErrNoRows", err)
	}
}

func TestWorkItemDurationCoefficient(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows([]string{"s"}).AddRow("3.5"))
	r := NewWorkItemRepo(m)
	d, err := r.DurationCoefficient(context.Background(), 9)
	if err != nil {
		t.Fatalf("DurationCoefficient: %v", err)
	}
	if d == nil || !d.Equal(decimal.RequireFromString("3.5")) {
		t.Errorf("d = %v", d)
	}
}

func TestWorkItemDurationCoefficientNull(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows([]string{"s"}).AddRow(nil))
	r := NewWorkItemRepo(m)
	d, err := r.DurationCoefficient(context.Background(), 9)
	if err != nil {
		t.Fatalf("DurationCoefficient: %v", err)
	}
	if d != nil {
		t.Errorf("d = %v want nil", d)
	}
}

func TestWorkItemCreate(t *testing.T) {
	m := newPool(t)
	level := "2.1.1.1"
	dur := decimal.NewFromInt(3)
	m.ExpectQuery(`INSERT INTO work_items`).
		WithArgs(int32(1), models.CategoryRoof, "Rangka atap", "10", "m2", "522451", "5224510",
			models.MethodAHSP, &level, (*string)(nil), (*int32)(nil), nil, "3").
		WillReturnRows(pgxmock.NewRows(workItemRow()).
			AddRow(int32(2), int32(1), models.CategoryRoof, "Rangka atap", "10", "m2", "522451", "5224510",
				models.MethodAHSP, "2.1.1.1", nil, nil, nil, "3", "30", time.Now(), time.Now()))
	r := NewWorkItemRepo(m)
	it, err := r.Create(context.Background(), nil, CreateWorkItemInput{
		ProjectID: 1, Category: models.CategoryRoof, Description: "Rangka atap",
		Volume: decimal.NewFromInt(10), Unit: "m2", UnitPrice: decimal.NewFromInt(522451),
		TotalCost: decimal.NewFromInt(5224510), CalculationMethod: models.MethodAHSP,
		Level: &level, Duration: &dur,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if it.ID != 2 {
		t.Errorf("id = %d", it.ID)
	}
}

func TestWorkItemUpdate(t *testing.T) {
	m := newPool(t)
	vol := decimal.NewFromInt(20)
	desc := "Updated"
	m.ExpectQuery(`UPDATE work_items SET`).
		WithArgs(int32(2), "20", nil, nil, &desc, (*string)(nil), (*string)(nil), (*string)(nil), (*models.CalculationMethod)(nil), nil).
		WillReturnRows(pgxmock.NewRows(workItemRow()).
			AddRow(int32(2), int32(1), models.CategoryRoof, "Updated", "20", "m2", "522451", "10449020",
				models.MethodAHSP, nil, nil, nil, nil, nil, nil, time.Now(), time.Now()))
	r := NewWorkItemRepo(m)
	it, err := r.Update(context.Background(), 2, UpdateWorkItemInput{Volume: &vol, Description: &desc})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !it.Volume.Equal(decimal.NewFromInt(20)) {
		t.Errorf("volume = %s", it.Volume)
	}
}

func TestWorkItemDelete(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`UPDATE work_items SET "deletedAt"`).WithArgs(int32(2)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	r := NewWorkItemRepo(m)
	if err := r.Delete(context.Background(), 2); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestWorkItemSetProgressClamps(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`UPDATE work_items SET progress`).WithArgs(int32(2), 100).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	r := NewWorkItemRepo(m)
	if err := r.SetProgress(context.Background(), 2, 250); err != nil {
		t.Fatalf("SetProgress: %v", err)
	}
}

func TestWorkItemRecordProgressHappyPath(t *testing.T) {
	m := newPool(t)
	note := "mulai ngecor"
	m.ExpectBegin()
	m.ExpectQuery(`SELECT progress FROM work_items`).WithArgs(int32(2)).
		WillReturnRows(pgxmock.NewRows([]string{"progress"}).AddRow(25))
	m.ExpectExec(`UPDATE work_items SET progress`).WithArgs(int32(2), 50).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	m.ExpectQuery(`INSERT INTO work_item_progress_logs`).
		WithArgs(int32(2), 50, &note).
		WillReturnRows(pgxmock.NewRows([]string{"id", "createdAt"}).AddRow(int32(7), time.Now()))
	m.ExpectCommit()
	r := NewWorkItemRepo(m)
	log, err := r.RecordProgress(context.Background(), 2, 50, &note)
	if err != nil {
		t.Fatalf("RecordProgress: %v", err)
	}
	if log == nil || log.ID != 7 || log.Progress != 50 {
		t.Errorf("log = %+v", log)
	}
	if log.Note == nil || *log.Note != note {
		t.Errorf("note = %v", log.Note)
	}
}

func TestWorkItemRecordProgressNotIncreasing(t *testing.T) {
	m := newPool(t)
	m.ExpectBegin()
	m.ExpectQuery(`SELECT progress FROM work_items`).WithArgs(int32(2)).
		WillReturnRows(pgxmock.NewRows([]string{"progress"}).AddRow(50))
	r := NewWorkItemRepo(m)
	if _, err := r.RecordProgress(context.Background(), 2, 25, nil); !errors.Is(err, ErrProgressNotIncreasing) {
		t.Fatalf("err = %v want ErrProgressNotIncreasing", err)
	}
}

func TestWorkItemListProgressLogs(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_item_progress_logs WHERE`).WithArgs(int32(2)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "workItemId", "progress", "note", "createdAt"}).
			AddRow(int32(1), int32(2), 10, "start", time.Now()).
			AddRow(int32(2), int32(2), 40, nil, time.Now()))
	r := NewWorkItemRepo(m)
	logs, err := r.ListProgressLogs(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListProgressLogs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("logs = %d", len(logs))
	}
	if logs[0].Note == nil || *logs[0].Note != "start" {
		t.Errorf("note0 = %v", logs[0].Note)
	}
	if logs[1].Note != nil {
		t.Errorf("note1 = %v", logs[1].Note)
	}
}

func TestWorkItemCreateDetail(t *testing.T) {
	m := newPool(t)
	src := "M.12"
	m.ExpectExec(`INSERT INTO work_item_details`).
		WithArgs(int32(2), (*int32)(nil), int32Ptr(9), "Semen", "kg", "420", "1750", "735000", models.ComponentMaterial, &src).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	r := NewWorkItemRepo(m)
	err := r.CreateDetail(context.Background(), nil, CreateDetailInput{
		WorkItemID: 2, AnalysisMasterID: int32Ptr(9), Name: "Semen", Unit: "kg",
		Coefficient: decimal.NewFromInt(420), UnitPrice: decimal.NewFromInt(1750),
		TotalCost: decimal.NewFromInt(735000), Type: models.ComponentMaterial, SourceCode: &src,
	})
	if err != nil {
		t.Fatalf("CreateDetail: %v", err)
	}
}

func int32Ptr(v int32) *int32 { return &v }

func TestWorkItemListDetails(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM work_item_details WHERE`).WithArgs(int32(2)).
		WillReturnRows(pgxmock.NewRows(detailRow()).
			AddRow(int32(1), int32(2), int32(20), nil, "Pekerja", "OH", "0.2", "100000", "20000", models.ComponentLabor, time.Now(), nil))
	r := NewWorkItemRepo(m)
	ds, err := r.ListWorkItemDetail(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListWorkItemDetail: %v", err)
	}
	if len(ds) != 1 {
		t.Fatalf("details = %d", len(ds))
	}
	if ds[0].PriceMasterID == nil || *ds[0].PriceMasterID != 20 {
		t.Errorf("priceMasterId = %v", ds[0].PriceMasterID)
	}
	if ds[0].AnalysisMasterID != nil {
		t.Errorf("analysisMasterId = %v", ds[0].AnalysisMasterID)
	}
	if !ds[0].Coefficient.Equal(decimal.RequireFromString("0.2")) {
		t.Errorf("coefficient = %s", ds[0].Coefficient)
	}
	if ds[0].SourceCode != nil {
		t.Errorf("sourceCode = %v", ds[0].SourceCode)
	}
}

func TestWorkItemDeleteDetails(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`DELETE FROM work_item_details`).WithArgs(int32(2)).
		WillReturnResult(pgxmock.NewResult("DELETE", 3))
	r := NewWorkItemRepo(m)
	if err := r.DeleteDetails(context.Background(), nil, 2); err != nil {
		t.Fatalf("DeleteDetails: %v", err)
	}
}

func TestWorkItemSetTotal(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`UPDATE work_items SET "unitPrice"`).
		WithArgs(int32(2), "999000", "19980000").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	r := NewWorkItemRepo(m)
	if err := r.SetTotal(context.Background(), nil, 2, decimal.NewFromInt(999000), decimal.NewFromInt(19980000)); err != nil {
		t.Fatalf("SetTotal: %v", err)
	}
}
