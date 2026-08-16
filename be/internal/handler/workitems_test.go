package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
	"github.com/halora-land/halora-be/service"
)

const workItemsSQL = `FROM work_items WHERE "deletedAt" IS NULL`

func workItemListRows(rows ...[]any) *pgxmock.Rows {
	rr := pgxmock.NewRows([]string{"id", "projectId", "category", "description", "volume", "unit",
		"unitPrice", "totalCost", "calculationMethod", "level", "type", "analysisMasterId",
		"basePrice", "duration", "totalDuration", "createdAt", "updatedAt"})
	for _, r := range rows {
		rr.AddRow(r...)
	}
	return rr
}

func wiRow(id int32, pid int32, cat models.WorkCategory, desc, cost string) []any {
	return []any{id, pid, cat, desc, "1", "bh", "50000", cost, models.MethodManual, nil, nil, nil, nil,
		"8", "8", time.Now(), time.Now()}
}

func outRaw(t *testing.T, w *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	var out []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestWorkItemListMergesWeightsWithProjectFilter(t *testing.T) {
	m := newPool(t)
	pid := int32(3)
	m.ExpectQuery(workItemsSQL + ` AND "projectId" = \$1`).WithArgs(pid).
		WillReturnRows(workItemListRows(
			wiRow(1, pid, models.CategoryFoundation, "Pondasi", "40000"),
			wiRow(2, pid, models.CategoryFoundation, "Besi", "60000")))
	// progress service weights: 40k/100k = 40, 60k/100k = 60
	m.ExpectQuery(`SELECT id, category, description, trim_scale\(ROUND\(volume, 2\)\)::text`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows([]string{"id", "category", "description", "volume", "unit",
			"totalCost", "progress", "duration"}).
			AddRow(int32(1), "foundation", "Pondasi", "1", "bh", "40000", 50, "8").
			AddRow(int32(2), "foundation", "Besi", "1", "bh", "60000", 100, "8"))

	h := NewWorkItemHandler(m, repository.NewWorkItemRepo(m),
		service.NewSnapshotService(m, repository.NewWorkItemRepo(m), nil, nil),
		service.NewProgressService(m), nil)
	w := doReq(t, h.List, http.MethodGet, "/api/v1/work-items?projectId=3", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
	var out []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("items = %d", len(out))
	}
	if out[0]["weight"].(float64) != 40 || out[1]["weight"].(float64) != 60 {
		t.Errorf("weights = %v, %v; want 40, 60", out[0]["weight"], out[1]["weight"])
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestWorkItemListWithoutProjectFilterKeepsWeightsZero(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(workItemsSQL + ` ORDER BY id ASC`).WithArgs().
		WillReturnRows(workItemListRows(wiRow(1, 3, models.CategoryFoundation, "Pondasi", "40000")))
	h := NewWorkItemHandler(m, repository.NewWorkItemRepo(m), nil, service.NewProgressService(m), nil)
	w := doReq(t, h.List, http.MethodGet, "/api/v1/work-items", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if _, ok := outRaw(t, w)[0]["weight"]; ok {
		t.Error("weight present without project filter, want omitted")
	}
}

func TestWorkItemListProjectFilterNoMatches(t *testing.T) {
	m := newPool(t)
	pid := int32(9)
	m.ExpectQuery(workItemsSQL + ` AND "projectId" = \$1`).WithArgs(pid).
		WillReturnRows(workItemListRows())
	m.ExpectQuery(`SELECT id, category, description, trim_scale\(ROUND\(volume, 2\)\)::text`).WithArgs(pid).
		WillReturnRows(pgxmock.NewRows([]string{"id", "category", "description", "volume", "unit",
			"totalCost", "progress", "duration"}))
	h := NewWorkItemHandler(m, repository.NewWorkItemRepo(m), nil, service.NewProgressService(m), nil)
	w := doReq(t, h.List, http.MethodGet, "/api/v1/work-items?projectId=9", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if got := outRaw(t, w); len(got) != 0 {
		t.Errorf("items = %v, want empty", got)
	}
}

func TestWorkItemListCategorySearchFilter(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(workItemsSQL+` AND category = \$1 AND "description" ILIKE \$2`).
		WithArgs(string(models.CategoryFoundation), "%besi%").
		WillReturnRows(workItemListRows(wiRow(1, 3, models.CategoryFoundation, "Besi", "40000")))
	h := NewWorkItemHandler(m, repository.NewWorkItemRepo(m), nil, nil, nil)
	w := doReq(t, h.List, http.MethodGet,
		"/api/v1/work-items?category=foundation&search=besi", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if got := outRaw(t, w); len(got) != 1 {
		t.Errorf("items = %d, want 1", len(got))
	}
}

func TestWorkItemListQueryError(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(workItemsSQL).WithArgs().WillReturnError(errors.New("boom"))
	h := NewWorkItemHandler(m, repository.NewWorkItemRepo(m), nil, service.NewProgressService(m), nil)
	w := doReq(t, h.List, http.MethodGet, "/api/v1/work-items", "", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
