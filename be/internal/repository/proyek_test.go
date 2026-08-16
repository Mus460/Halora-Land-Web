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

func projectRow() []string {
	return []string{"id", "userId", "name", "location", "type", "isPitching", "isDone", "contractValue",
		"buildingArea", "timelineMonths", "timelineDays", "createdAt", "updatedAt"}
}

func TestProjectListForUser(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM projects p LEFT JOIN project_team`).WithArgs(int32(4)).
		WillReturnRows(pgxmock.NewRows(projectRow()).
			AddRow(int32(1), int32(4), "Ruko 2 Lantai", "Jakarta", models.ProjectTypeBuilding, false, false, "2500000000", nil, 6, 0, time.Now(), time.Now()).
			AddRow(int32(2), int32(9), "Rumah 70", nil, models.ProjectTypeBuilding, true, false, nil, nil, 3, 15, time.Now(), time.Now()))
	r := NewProjectRepo(m)
	ps, err := r.List(context.Background(), ListProjectFilter{UserID: 4})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ps) != 2 {
		t.Fatalf("projects = %d", len(ps))
	}
	if ps[0].Name != "Ruko 2 Lantai" || ps[0].Location == nil || *ps[0].Location != "Jakarta" {
		t.Errorf("p0 = %+v", ps[0])
	}
	if ps[1].ContractValue != nil || ps[1].IsPitching != true {
		t.Errorf("p1 = %+v", ps[1])
	}
	if ps[0].ContractValue == nil || !ps[0].ContractValue.Equal(decimal.NewFromInt(2500000000)) {
		t.Errorf("contractValue = %v", ps[0].ContractValue)
	}
}

func TestProjectListAdminWithSearchAndType(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM projects WHERE "deletedAt" IS NULL`).
		WithArgs("%ruko%", "building").
		WillReturnRows(pgxmock.NewRows(projectRow()).AddRow(int32(1), int32(4), "Ruko", "Jakarta", "building", false, false, nil, nil, 6, 0, time.Now(), time.Now()))
	r := NewProjectRepo(m)
	ps, err := r.List(context.Background(), ListProjectFilter{IsAdmin: true, Search: "ruko", Type: "building"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ps) != 1 || ps[0].Name != "Ruko" {
		t.Fatalf("projects = %+v", ps)
	}
}

func TestProjectGet(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM projects WHERE id =`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(projectRow()).
			AddRow(int32(1), int32(4), "Ruko", "Jakarta", "building", false, false, "2500000000", nil, 6, 0, time.Now(), time.Now()))
	r := NewProjectRepo(m)
	p, err := r.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.UserID != 4 {
		t.Errorf("userId = %d", p.UserID)
	}
}

func TestProjectGetNotFound(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM projects WHERE id =`).WithArgs(int32(999)).WillReturnError(pgx.ErrNoRows)
	r := NewProjectRepo(m)
	if _, err := r.Get(context.Background(), 999); err == nil {
		t.Fatal("expected ErrNoRows")
	}
}

func TestProjectGetDetail(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM projects WHERE id =`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(projectRow()).
			AddRow(int32(1), int32(4), "Ruko", "Jakarta", "building", false, false, "2500000000", nil, 6, 0, time.Now(), time.Now()))
	m.ExpectQuery(`FROM users WHERE id =`).WithArgs(int32(4)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email"}).AddRow(int32(4), "Budi", "budi@x.id"))
	m.ExpectQuery(`FROM project_team tp`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "role", "userId", "fullName", "email"}).
			AddRow(int32(10), models.TeamRoleEditor, int32(9), "Sari", "sari@x.id"))
	m.ExpectQuery(`ORDER BY id DESC LIMIT 10`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "description", "volume", "unit", "unitPrice", "totalCost", "category"}).
			AddRow(int32(5), "Galian", "12", "m3", "50000", "600000", models.CategoryPreparation))
	m.ExpectQuery(`FROM work_items WHERE`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"c"}).AddRow(int32(3)))
	m.ExpectQuery(`FROM invoices WHERE`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"c"}).AddRow(int32(2)))

	r := NewProjectRepo(m)
	d, err := r.GetDetail(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if d.User.FullName != "Budi" {
		t.Errorf("owner = %+v", d.User)
	}
	if len(d.ProjectTeam) != 1 || d.ProjectTeam[0].Role != models.TeamRoleEditor {
		t.Errorf("team = %+v", d.ProjectTeam)
	}
	if len(d.WorkItem) != 1 || !d.WorkItem[0].Volume.Equal(decimal.NewFromInt(12)) {
		t.Errorf("workItems = %+v", d.WorkItem)
	}
	if d.Count.WorkItem != 3 || d.Count.Invoice != 2 {
		t.Errorf("count = %+v", d.Count)
	}
}

func TestProjectCreate(t *testing.T) {
	m := newPool(t)
	loc := "Depok"
	cv := decimal.NewFromInt(1500000000)
	m.ExpectQuery(`INSERT INTO projects`).
		WithArgs(int32(4), "Rumah", &loc, models.ProjectTypeBuilding, false, true, "1500000000", nil, 4, 10).
		WillReturnRows(pgxmock.NewRows(projectRow()).
			AddRow(int32(7), int32(4), "Rumah", "Depok", "building", false, true, "1500000000", nil, 4, 10, time.Now(), time.Now()))
	r := NewProjectRepo(m)
	p, err := r.Create(context.Background(), CreateProjectInput{
		UserID: 4, Name: "Rumah", Location: &loc, Type: models.ProjectTypeBuilding,
		IsPitching: false, IsDone: true, ContractValue: &cv, TimelineMonths: 4, TimelineDays: 10,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID != 7 || !p.ContractValue.Equal(decimal.NewFromInt(1500000000)) {
		t.Errorf("p = %+v", p)
	}
}

func TestProjectImportBOQ(t *testing.T) {
	m := newPool(t)
	loc := "Bogor"
	m.ExpectBegin()
	m.ExpectQuery(`INSERT INTO projects`).
		WithArgs(int32(4), "Gedung", &loc, models.ProjectTypeBuilding, false, false, nil, nil, 0, 0).
		WillReturnRows(pgxmock.NewRows(projectRow()).
			AddRow(int32(8), int32(4), "Gedung", "Bogor", "building", false, false, nil, nil, 0, 0, time.Now(), time.Now()))
	m.ExpectExec(`INSERT INTO work_items`).
		WithArgs(int32(8), models.CategoryWall, "Pasang bata", "20", "m2", "120000", "2400000").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectExec(`INSERT INTO work_items`).
		WithArgs(int32(8), models.CategoryRoof, "Atap", "30", "m2", "200000", "6000000").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	m.ExpectCommit()

	r := NewProjectRepo(m)
	p, err := r.ImportBOQ(context.Background(), CreateProjectInput{UserID: 4, Name: "Gedung", Location: &loc, Type: models.ProjectTypeBuilding},
		[]ImportedWorkItem{
			{Category: models.CategoryWall, Description: "Pasang bata", Volume: decimal.NewFromInt(20), Unit: "m2", UnitPrice: decimal.NewFromInt(120000), TotalCost: decimal.NewFromInt(2400000)},
			{Category: models.CategoryRoof, Description: "Atap", Volume: decimal.NewFromInt(30), Unit: "m2", UnitPrice: decimal.NewFromInt(200000), TotalCost: decimal.NewFromInt(6000000)},
		})
	if err != nil {
		t.Fatalf("ImportBOQ: %v", err)
	}
	if p.ID != 8 {
		t.Errorf("id = %d", p.ID)
	}
}

func TestProjectImportBOQSkipsEmptyDescriptions(t *testing.T) {
	m := newPool(t)
	m.ExpectBegin()
	m.ExpectQuery(`INSERT INTO projects`).
		WithArgs(int32(4), "Gedung", (*string)(nil), models.ProjectTypeBuilding, false, false, nil, nil, 0, 0).
		WillReturnRows(pgxmock.NewRows(projectRow()).
			AddRow(int32(8), int32(4), "Gedung", nil, "building", false, false, nil, nil, 0, 0, time.Now(), time.Now()))
	m.ExpectCommit()
	r := NewProjectRepo(m)
	_, err := r.ImportBOQ(context.Background(), CreateProjectInput{UserID: 4, Name: "Gedung", Type: models.ProjectTypeBuilding},
		[]ImportedWorkItem{{Description: "", Volume: decimal.NewFromInt(1)}})
	if err != nil {
		t.Fatalf("ImportBOQ: %v", err)
	}
}

func TestProjectUpdate(t *testing.T) {
	m := newPool(t)
	name := "Ruko Baru"
	m.ExpectQuery(`UPDATE projects SET`).
		WithArgs(int32(1), &name, (*string)(nil), (*models.ProjectType)(nil), (*bool)(nil), (*bool)(nil), nil, nil, (*int)(nil), (*int)(nil)).
		WillReturnRows(pgxmock.NewRows(projectRow()).
			AddRow(int32(1), int32(4), "Ruko Baru", "Jakarta", "building", false, false, nil, nil, 6, 0, time.Now(), time.Now()))
	r := NewProjectRepo(m)
	p, err := r.Update(context.Background(), 1, UpdateProjectInput{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if p.Name != "Ruko Baru" {
		t.Errorf("name = %q", p.Name)
	}
}

func TestProjectDelete(t *testing.T) {
	m := newPool(t)
	m.ExpectBegin()
	m.ExpectExec(`UPDATE projects SET "deletedAt"`).WithArgs(int32(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	m.ExpectExec(`UPDATE work_items SET "deletedAt"`).WithArgs(int32(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 5))
	m.ExpectCommit()
	r := NewProjectRepo(m)
	if err := r.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestProjectSummary(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM projects WHERE id =`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows(projectRow()).
			AddRow(int32(1), int32(4), "Ruko", "Jakarta", "building", false, false, "2500000000", nil, 6, 0, time.Now(), time.Now()))
	r := NewProjectRepo(m)
	s, err := r.Summary(context.Background(), 1)
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if s.ID != 1 || s.Name != "Ruko" {
		t.Errorf("s = %+v", s)
	}
	if s.ContractValue == nil || !s.ContractValue.Equal(decimal.NewFromInt(2500000000)) {
		t.Errorf("contract = %v", s.ContractValue)
	}
}
