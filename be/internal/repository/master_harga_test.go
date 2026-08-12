package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

func priceRow() []string {
	return []string{"id", "name", "unit", "price", "type", "isGlobal", "userId", "ahspCode", "isSystem", "createdAt", "updatedAt"}
}

func TestPriceMasterListCaches(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM price_masters WHERE`).WithArgs(int32(4), "material").
		WillReturnRows(pgxmock.NewRows(priceRow()).
			AddRow(int32(1), "Semen", "kg", "1750", models.ComponentMaterial, true, nil, "M.01", true, time.Now(), time.Now()))
	r := NewPriceMasterRepo(m)
	list, err := r.List(context.Background(), ListPriceMasterFilter{UserID: 4, Type: "material"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || !list[0].Price.Equal(decimal.NewFromInt(1750)) {
		t.Fatalf("list = %+v", list)
	}
	if list[0].AHSPCode == nil || *list[0].AHSPCode != "M.01" {
		t.Errorf("ahspCode = %v", list[0].AHSPCode)
	}
	again, err := r.List(context.Background(), ListPriceMasterFilter{UserID: 4, Type: "material"})
	if err != nil || len(again) != 1 {
		t.Fatalf("cached List: %v (%d)", err, len(again))
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestPriceMasterGet(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM price_masters WHERE id =`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(priceRow()).
			AddRow(int32(1), "Semen", "kg", "1750", models.ComponentMaterial, true, nil, "M.01", true, time.Now(), time.Now()))
	r := NewPriceMasterRepo(m)
	p, err := r.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.Name != "Semen" || p.Type != models.ComponentMaterial {
		t.Errorf("p = %+v", p)
	}
}

func TestPriceMasterGetNoRows(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM price_masters WHERE id =`).WithArgs(int32(1)).WillReturnError(pgx.ErrNoRows)
	r := NewPriceMasterRepo(m)
	if _, err := r.Get(context.Background(), 1); err == nil {
		t.Fatal("expected ErrNoRows")
	}
}

func TestPriceMasterCreate(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`INSERT INTO price_masters`).
		WithArgs("Besi 10", "batang", "125000", models.ComponentMaterial, true, (*int32)(nil), false).
		WillReturnRows(pgxmock.NewRows(priceRow()).
			AddRow(int32(2), "Besi 10", "batang", "125000", models.ComponentMaterial, true, nil, nil, false, time.Now(), time.Now()))
	r := NewPriceMasterRepo(m)
	p, err := r.Create(context.Background(), CreatePriceMasterInput{
		Name: "Besi 10", Unit: "batang", Price: decimal.NewFromInt(125000),
		Type: models.ComponentMaterial, IsGlobal: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID != 2 {
		t.Errorf("id = %d", p.ID)
	}
}

func TestPriceMasterUpdate(t *testing.T) {
	m := newPool(t)
	name := "Semen 50kg"
	price := decimal.NewFromInt(1900)
	m.ExpectQuery(`UPDATE price_masters SET`).
		WithArgs(int32(1), &name, (*string)(nil), "1900", (*models.ComponentType)(nil)).
		WillReturnRows(pgxmock.NewRows(priceRow()).
			AddRow(int32(1), "Semen 50kg", "kg", "1900", models.ComponentMaterial, true, nil, "M.01", true, time.Now(), time.Now()))
	r := NewPriceMasterRepo(m)
	p, err := r.Update(context.Background(), 1, UpdatePriceMasterInput{Name: &name, Price: &price})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !p.Price.Equal(decimal.NewFromInt(1900)) {
		t.Errorf("price = %s", p.Price)
	}
}

func TestPriceMasterDelete(t *testing.T) {
	m := newPool(t)
	m.ExpectExec(`UPDATE price_masters SET "deletedAt"`).WithArgs(int32(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	r := NewPriceMasterRepo(m)
	if err := r.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestPriceMasterGetMany(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM price_masters WHERE id = ANY`).WithArgs([]int32{1, 2}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "price"}).
			AddRow(int32(1), "100").
			AddRow(int32(2), "250.5"))
	r := NewPriceMasterRepo(m)
	got, err := r.GetMany(context.Background(), []int32{1, 2})
	if err != nil {
		t.Fatalf("GetMany: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got = %v", got)
	}
	if !got[1].Equal(decimal.RequireFromString("100")) || !got[2].Equal(decimal.RequireFromString("250.5")) {
		t.Errorf("got = %v", got)
	}
}

func TestPriceMasterGetManyEmpty(t *testing.T) {
	r := NewPriceMasterRepo(newPool(t))
	got, err := r.GetMany(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMany: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got = %v", got)
	}
}
