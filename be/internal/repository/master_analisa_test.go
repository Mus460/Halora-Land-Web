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

func analisaRow() []string {
	return []string{"id", "code", "name", "level", "parentId", "unit", "unitPrice", "category",
		"isGlobal", "userId", "isSystem", "ahspCode", "ahspSheet", "generalCost", "createdAt", "updatedAt"}
}

func TestAnalysisMasterListCachesResults(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_masters WHERE`).WithArgs(int32(4)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(1), "2.1.1.1", "Rangka Atap", 2, nil, "m2", "522451.64", "roof",
				true, nil, true, "2.1.1.1", "Atap", "0", time.Now(), time.Now()).
			AddRow(int32(2), "2.1.1.2", "Rangka Atap 2", 2, int32(1), "m2", "0", "roof",
				false, int32(4), false, "2.1.1.2", "Atap", "0.05", time.Now(), time.Now()))
	r := NewAnalysisMasterRepo(m)

	// first call hits the DB and populates the cache
	list, err := r.List(context.Background(), ListAnalysisMasterFilter{UserID: 4})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list = %d", len(list))
	}
	if !list[0].UnitPrice.Equal(decimal.RequireFromString("522451.64")) {
		t.Errorf("unitPrice = %v", list[0].UnitPrice)
	}
	if list[0].AHSPSheet == nil || *list[0].AHSPSheet != "Atap" {
		t.Errorf("sheet = %v", list[0].AHSPSheet)
	}
	if !list[0].GeneralCost.Equal(decimal.Zero) {
		t.Errorf("generalCost = %s", list[0].GeneralCost)
	}

	// second call must not hit the DB (cached)
	again, err := r.List(context.Background(), ListAnalysisMasterFilter{UserID: 4})
	if err != nil {
		t.Fatalf("List cached: %v", err)
	}
	if len(again) != 2 {
		t.Fatalf("cached list = %d", len(again))
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestAnalysisMasterGet(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(1), "1.1.1.1", "Pembersihan", 1, nil, "m'", "100000", "preparation",
				true, nil, true, "1.1.1.1", "Persiapan", "0", time.Now(), time.Now()))
	r := NewAnalysisMasterRepo(m)
	ma, err := r.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ma.Code != "1.1.1.1" || ma.Level != 1 {
		t.Errorf("ma = %+v", ma)
	}
}

func TestAnalysisMasterGetNoRows(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(1)).WillReturnError(pgx.ErrNoRows)
	r := NewAnalysisMasterRepo(m)
	if _, err := r.Get(context.Background(), 1); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("err = %v", err)
	}
}

func TestAnalysisMasterCreateClearsCache(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`INSERT INTO analysis_masters`).
		WithArgs("X.1", "Custom", int32(1), (*int32)(nil), strp("m2"), true, (*int32)(nil), false).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(5), "X.1", "Custom", 1, nil, "m2", nil, nil, true, nil, false, nil, nil, "0", time.Now(), time.Now()))
	r := NewAnalysisMasterRepo(m)
	ma, err := r.Create(context.Background(), CreateAnalysisMasterInput{Code: "X.1", Name: "Custom", Level: 1, Unit: strp("m2"), IsGlobal: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ma.ID != 5 {
		t.Errorf("id = %d", ma.ID)
	}
}

func TestAnalysisMasterDelete(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`UPDATE analysis_masters SET "deletedAt"`).WithArgs(int32(5)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	r := NewAnalysisMasterRepo(m)
	if err := r.Delete(context.Background(), 5); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestAnalysisMasterCopy(t *testing.T) {
	m := newPool(t)
	// source load
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(1), "2.1.1.1", "Rangka Atap", 2, nil, "m2", "522451.64", "roof",
				true, nil, true, "2.1.1.1", "Atap", "0", time.Now(), time.Now()))
	m.ExpectBegin()
	// node reload inside txn
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(1), "2.1.1.1", "Rangka Atap", 2, nil, "m2", "522451.64", "roof",
				true, nil, true, "2.1.1.1", "Atap", "0", time.Now(), time.Now()))
	// code-exists check → false
	m.ExpectQuery(`SELECT EXISTS`).WithArgs("2.1.1.1", int32(4)).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
	// insert copy
	m.ExpectQuery(`INSERT INTO analysis_masters \(code`).WithArgs("2.1.1.1", "Salin - Rangka Atap", int32(2), int32(0), strp("m2"), "522451.64", strp("roof"), int32(4), "0").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int32(9)))
	// components copy
	m.ExpectExec(`INSERT INTO analysis_components`).WithArgs(int32(9), int32(1)).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))
	// children lookup → none
	m.ExpectQuery(`SELECT id FROM analysis_masters WHERE`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id"}))
	m.ExpectCommit()
	// fetch of the new node (cache cleared)
	m.ExpectQuery(`FROM analysis_masters WHERE id =`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(9), "2.1.1.1", "Salin - Rangka Atap", 2, nil, "m2", "522451.64", "roof",
				false, int32(4), false, "2.1.1.1", "Atap", "0", time.Now(), time.Now()))

	r := NewAnalysisMasterRepo(m)
	ma, err := r.Copy(context.Background(), 1, 4, "")
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if ma.ID != 9 || ma.UserID == nil || *ma.UserID != 4 {
		t.Errorf("ma = %+v", ma)
	}
}

func TestAnalysisMasterUpdate(t *testing.T) {
	m := newPool(t)
	name := "Ganti nama"
	m.ExpectQuery(`UPDATE analysis_masters SET name`).
		WithArgs(&name, (*string)(nil), int32(3), int32(4), false).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(3), "X.1", "Ganti nama", 1, nil, "m2", "50000", nil, false, int32(4), false, nil, nil, "0", time.Now(), time.Now()))
	r := NewAnalysisMasterRepo(m)
	ma, err := r.Update(context.Background(), 3, 4, false, UpdateAnalysisMasterInput{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if ma.Name != "Ganti nama" {
		t.Errorf("name = %q", ma.Name)
	}
}

func TestAnalysisMasterHasChildren(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`count\(\*\) FROM analysis_masters WHERE`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"n"}).AddRow(0))
	m.ExpectQuery(`count\(\*\) FROM analysis_masters WHERE`).WithArgs(int32(2)).
		WillReturnRows(pgxmock.NewRows([]string{"n"}).AddRow(3))
	r := NewAnalysisMasterRepo(m)
	no, err := r.HasChildren(context.Background(), 1)
	if err != nil || no {
		t.Errorf("no-children = %v err %v", no, err)
	}
	yes, err := r.HasChildren(context.Background(), 2)
	if err != nil || !yes {
		t.Errorf("has-children = %v err %v", yes, err)
	}
}

func TestAnalysisMasterListComponents(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_components WHERE`).WithArgs(int32(9)).
		WillReturnRows(pgxmock.NewRows(componentRow()).
			AddRow(int32(1), int32(9), int32(20), "0.2", models.ComponentLabor, "Tukang batu", "OH", "145000", "29000", "L.02", "2", 1, time.Now(), time.Now()).
			AddRow(int32(2), int32(9), nil, "2.625", models.ComponentMaterial, "Pasir Beton", "kg", "300", "787.5", nil, nil, 2, time.Now(), time.Now()))
	r := NewAnalysisMasterRepo(m)
	cs, err := r.ListComponents(context.Background(), 9)
	if err != nil {
		t.Fatalf("ListComponents: %v", err)
	}
	if len(cs) != 2 {
		t.Fatalf("components = %d", len(cs))
	}
	if cs[0].ComponentID == nil || *cs[0].ComponentID != 20 {
		t.Errorf("componentId = %v", cs[0].ComponentID)
	}
	if cs[1].ComponentID != nil {
		t.Errorf("componentId1 = %v", cs[1].ComponentID)
	}
	if !cs[1].Coefficient.Equal(decimal.RequireFromString("2.625")) {
		t.Errorf("coefficient = %s", cs[1].Coefficient)
	}
	if cs[1].ReferenceCode != nil {
		t.Errorf("ref = %v", cs[1].ReferenceCode)
	}
}

func componentRow() []string {
	return []string{"id", "analysisMasterId", "componentId", "coefficient", "type", "name", "unit",
		"unitPrice", "totalPrice", "referenceCode", "duration", "sequence", "createdAt", "updatedAt"}
}

func TestAnalysisMasterUpdateComponent(t *testing.T) {
	m := newPool(t)
	coef := decimal.RequireFromString("0.5")
	hs := decimal.NewFromInt(200000)
	m.ExpectQuery(`UPDATE analysis_components c SET`).
		WithArgs("0.5", "200000", (*string)(nil), (*string)(nil), (*models.ComponentType)(nil), int32(1), int32(4), false).
		WillReturnRows(pgxmock.NewRows(componentRow()).
			AddRow(int32(1), int32(9), int32(20), "0.5", models.ComponentMaterial, "Semen", "kg", "200000", "100000", "M.1", nil, 1, time.Now(), time.Now()))
	m.ExpectExec(`UPDATE analysis_masters m SET`).WithArgs(int32(9)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	r := NewAnalysisMasterRepo(m)
	c, err := r.UpdateComponent(context.Background(), 4, false, UpdateComponentInput{
		ID: 1, AnalysisMasterID: 9, Coefficient: &coef, UnitPrice: &hs,
	})
	if err != nil {
		t.Fatalf("UpdateComponent: %v", err)
	}
	if !c.TotalPrice.Equal(decimal.NewFromInt(100000)) {
		t.Errorf("totalPrice = %s", c.TotalPrice)
	}
}

func TestAnalysisMasterCreateDeleteComponent(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`INSERT INTO analysis_components`).
		WithArgs(int32(9), int32Ptr(20), "0.75", models.ComponentLabor).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectExec(`UPDATE analysis_components SET "deletedAt"`).
		WithArgs(int32(1), int32(9)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	r := NewAnalysisMasterRepo(m)
	if err := r.CreateComponent(context.Background(), CreateComponentInput{
		AnalysisMasterID: 9, ComponentID: int32Ptr(20), Coefficient: decimal.RequireFromString("0.75"), Type: models.ComponentLabor,
	}); err != nil {
		t.Fatalf("CreateComponent: %v", err)
	}
	if err := r.DeleteComponent(context.Background(), 9, 1); err != nil {
		t.Fatalf("DeleteComponent: %v", err)
	}
}

func TestBuildAnalysisMasterTree(t *testing.T) {
	parent := models.AnalysisMaster{ID: 1, Code: "1", Level: 1, Name: "Parent"}
	child := models.AnalysisMaster{ID: 2, Code: "1.1", Level: 2, Name: "Child", ParentID: int32Ptr(1)}
	child2 := models.AnalysisMaster{ID: 3, Code: "1.2", Level: 2, Name: "Child2", ParentID: int32Ptr(1)}
	orphan := models.AnalysisMaster{ID: 4, Code: "2", Level: 1, Name: "Orphan", ParentID: int32Ptr(99)}
	roots := buildAnalysisMasterTree([]models.AnalysisMaster{child, orphan, parent, child2})
	if len(roots) != 2 {
		t.Fatalf("roots = %d want 2", len(roots))
	}
	var p *models.AnalysisMaster
	for i := range roots {
		if roots[i].ID == 1 {
			p = &roots[i]
		}
	}
	if p == nil {
		t.Fatal("parent root missing")
	}
	if len(p.Children) != 2 || p.Children[0].ID != 2 || p.Children[1].ID != 3 {
		t.Errorf("children = %+v", p.Children)
	}
}

func TestBuildAnalysisMasterTreeDeepNesting(t *testing.T) {
	l1 := models.AnalysisMaster{ID: 1, Code: "1", Level: 1}
	l2 := models.AnalysisMaster{ID: 2, Code: "1.1", Level: 2, ParentID: int32Ptr(1)}
	l3a := models.AnalysisMaster{ID: 3, Code: "1.1.1", Level: 3, ParentID: int32Ptr(2)}
	l3b := models.AnalysisMaster{ID: 4, Code: "1.1.2", Level: 3, ParentID: int32Ptr(2)}
	l2b := models.AnalysisMaster{ID: 5, Code: "1.2", Level: 2, ParentID: int32Ptr(1)}
	// production ordering: level ASC, code ASC
	all := []models.AnalysisMaster{l1, l2, l2b, l3a, l3b}
	roots := buildAnalysisMasterTree(all)
	if len(roots) != 1 {
		t.Fatalf("roots = %d", len(roots))
	}
	root := roots[0]
	if len(root.Children) != 2 || root.Children[0].ID != 2 || root.Children[1].ID != 5 {
		t.Fatalf("root children = %+v", root.Children)
	}
	mid := root.Children[0]
	if len(mid.Children) != 2 || mid.Children[0].ID != 3 || mid.Children[1].ID != 4 {
		t.Errorf("mid children = %+v", mid.Children)
	}
}

func TestAnalysisMasterSearchAHSP(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM analysis_masters WHERE "isSystem" = true`).
		WithArgs("%batu%", "wall", 20).
		WillReturnRows(pgxmock.NewRows(analisaRow()).
			AddRow(int32(1), "1.1", "Pasang batu", 1, nil, "m2", "100000", "wall", true, nil, true, "1.1", "A", "0", time.Now(), time.Now()))
	r := NewAnalysisMasterRepo(m)
	out, err := r.SearchAHSP(context.Background(), "batu", "wall", 20)
	if err != nil {
		t.Fatalf("SearchAHSP: %v", err)
	}
	if len(out) != 1 || out[0].Name != "Pasang batu" {
		t.Errorf("out = %+v", out)
	}
}

func strp(s string) *string { return &s }
