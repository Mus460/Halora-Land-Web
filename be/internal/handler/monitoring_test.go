package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/service"
)

const projectAccessSQL = `SELECT p."userId", tp.role`
const progressItemsSQL = `SELECT id, category, description, trim_scale\(ROUND\(volume, 2\)\)::text`
const lastUpdatedSQL = `DISTINCT ON \("workItemId"\)`

func monHandler(m pgxmock.PgxPoolIface) *MonitoringHandler {
	return NewMonitoringHandler(m, service.NewProgressService(m))
}

func adminCtx() context.Context {
	return auth.WithUser(context.Background(), &auth.AuthUser{UserID: 1, Role: models.RoleAdmin})
}

func expectProjectAccess(m pgxmock.PgxPoolIface, pid, uid int32, ownerID int32) {
	m.ExpectQuery(projectAccessSQL).WithArgs(pid, uid).
		WillReturnRows(pgxmock.NewRows([]string{"userId", "role"}).AddRow(ownerID, nil))
}

type monResp struct {
	Overall    int `json:"overall"`
	Monitoring []struct {
		Category string `json:"category"`
		Progress int    `json:"progress"`
		Items    []struct {
			ID          int32    `json:"id"`
			Description string   `json:"description"`
			Volume      string   `json:"volume"`
			Unit        string   `json:"unit"`
			Progress    int      `json:"progress"`
			Weight      float64  `json:"weight"`
			LastUpdated *string  `json:"lastUpdated"`
		} `json:"items"`
	} `json:"monitoring"`
}

func TestMonitoringListHappyPath(t *testing.T) {
	m := newPool(t)
	pid := int32(7)
	u := &auth.AuthUser{UserID: 1, Role: models.RoleUser}
	ctx := auth.WithUser(context.Background(), u)
	expectProjectAccess(m, pid, 1, 1)
	m.ExpectQuery(progressItemsSQL).WithArgs(pid).WillReturnRows(pgxmock.NewRows([]string{
		"id", "category", "description", "volume", "unit", "totalCost", "progress", "duration"}).
		AddRow(int32(1), "preparation", "Pembersihan", "10", "m2", "30000", 50, "8").
		AddRow(int32(2), "preparation", "Papan nama", "1", "bh", "30000", 0, "").
		AddRow(int32(3), "roof", "Atap", "20", "m2", "70000", 100, "8"))
	m.ExpectQuery(lastUpdatedSQL).WithArgs([]int32{1, 2, 3}).
		WillReturnRows(pgxmock.NewRows([]string{"workItemId", "createdAt"}).
			AddRow(int32(1), "2024-02-01T08:00:00Z").
			AddRow(int32(3), nil))

	w := doReq(t, monHandler(m).List, http.MethodGet, "/api/v1/monitoring?projectId=7", "", ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
	var out monResp
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	// weights: 30k/130k = 23.0769→23.08, 30k/130k, 70k/130k = 53.846→53.85
	if len(out.Monitoring) != 2 {
		t.Fatalf("categories = %d", len(out.Monitoring))
	}
	prep := out.Monitoring[0]
	if prep.Category != "preparation" || len(prep.Items) != 2 {
		t.Fatalf("prep = %+v", prep)
	}
	// category progress = (23.08×50 + 23.08×0)/46.16 ≈ 25
	if prep.Progress != 25 {
		t.Errorf("prep progress = %d, want 25", prep.Progress)
	}
	if out.Monitoring[1].Progress != 100 {
		t.Errorf("roof progress = %d, want 100", out.Monitoring[1].Progress)
	}
	// overall = 23.08×0.5 + 23.08×0 + 53.85×1.0 = 65.39 → 65
	if out.Overall != 65 {
		t.Errorf("overall = %d, want 65", out.Overall)
	}
	if prep.Items[0].Weight <= 23.07 || prep.Items[0].Weight >= 23.09 {
		t.Errorf("weight = %v, want ~23.08", prep.Items[0].Weight)
	}
	if out.Monitoring[1].Items[0].Weight <= 53.84 || out.Monitoring[1].Items[0].Weight >= 53.86 {
		t.Errorf("roof weight = %v, want ~53.85", out.Monitoring[1].Items[0].Weight)
	}
	if prep.Items[0].LastUpdated == nil || *prep.Items[0].LastUpdated != "2024-02-01T08:00:00Z" {
		t.Errorf("lastUpdated = %v", prep.Items[0].LastUpdated)
	}
	if prep.Items[1].LastUpdated != nil {
		t.Errorf("item without log lastUpdated = %v, want nil", prep.Items[1].LastUpdated)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestMonitoringListAdminSeesAllProjects(t *testing.T) {
	m := newPool(t)
	pid := int32(7)
	// Admin bypasses ownership; the access row scan still runs.
	m.ExpectQuery(projectAccessSQL).WithArgs(pid, int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"userId", "role"}).AddRow(int32(99), nil))
	m.ExpectQuery(progressItemsSQL).WithArgs(pid).WillReturnRows(pgxmock.NewRows([]string{
		"id", "category", "description", "volume", "unit", "totalCost", "progress", "duration"}).
		AddRow(int32(1), "preparation", "A", "1", "bh", "100", 0, "8"))
	m.ExpectQuery(lastUpdatedSQL).WithArgs([]int32{1}).
		WillReturnRows(pgxmock.NewRows([]string{"workItemId", "createdAt"}))

	w := doReq(t, monHandler(m).List, http.MethodGet, "/api/v1/monitoring?projectId=7", "", adminCtx())
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
	var out monResp
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Overall != 0 || len(out.Monitoring) != 1 {
		t.Errorf("overall/monitoring = %d/%d", out.Overall, len(out.Monitoring))
	}
}

func TestMonitoringListInvalidProjectID(t *testing.T) {
	m := newPool(t)
	h := monHandler(m)
	for _, pid := range []string{"abc", "12.5", "-3", "0", "99999999999999"} {
		w := doReq(t, h.List, http.MethodGet, "/api/v1/monitoring?projectId="+pid, "", adminCtx())
		if w.Code != http.StatusBadRequest {
			t.Errorf("projectId %q status = %d, want 400", pid, w.Code)
		}
	}
}

func TestMonitoringListProjectNotFound(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(projectAccessSQL).WithArgs(int32(404), int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"userId", "role"}))
	w := doReq(t, monHandler(m).List, http.MethodGet, "/api/v1/monitoring?projectId=404", "", adminCtx())
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestMonitoringListForbidden(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(projectAccessSQL).WithArgs(int32(7), int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"userId", "role"}).AddRow(int32(99), nil))
	u := &auth.AuthUser{UserID: 1, Role: models.RoleUser}
	w := doReq(t, monHandler(m).List, http.MethodGet, "/api/v1/monitoring?projectId=7", "",
		auth.WithUser(context.Background(), u))
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestMonitoringListItemsError(t *testing.T) {
	m := newPool(t)
	expectProjectAccess(m, 7, 1, 1)
	m.ExpectQuery(progressItemsSQL).WithArgs(int32(7)).WillReturnError(errors.New("boom"))
	w := doReq(t, monHandler(m).List, http.MethodGet, "/api/v1/monitoring?projectId=7", "", adminCtx())
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestMonitoringListEmptyProject(t *testing.T) {
	m := newPool(t)
	expectProjectAccess(m, 7, 1, 1)
	m.ExpectQuery(progressItemsSQL).WithArgs(int32(7)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "category", "description", "volume", "unit", "totalCost", "progress", "duration"}))
	w := doReq(t, monHandler(m).List, http.MethodGet, "/api/v1/monitoring?projectId=7", "", adminCtx())
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var out monResp
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Overall != 0 || len(out.Monitoring) != 0 {
		t.Errorf("overall/monitoring = %d/%d, want 0/empty", out.Overall, len(out.Monitoring))
	}
}

func TestMonitoringListLastUpdatedError(t *testing.T) {
	m := newPool(t)
	expectProjectAccess(m, 7, 1, 1)
	m.ExpectQuery(progressItemsSQL).WithArgs(int32(7)).WillReturnRows(pgxmock.NewRows([]string{
		"id", "category", "description", "volume", "unit", "totalCost", "progress", "duration"}).
		AddRow(int32(1), "preparation", "A", "1", "bh", "100", 0, "8"))
	m.ExpectQuery(lastUpdatedSQL).WithArgs([]int32{1}).WillReturnError(errors.New("boom"))
	w := doReq(t, monHandler(m).List, http.MethodGet, "/api/v1/monitoring?projectId=7", "", adminCtx())
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestMonitoringListAnonymous(t *testing.T) {
	m := newPool(t)
	h := monHandler(m)
	w := doReq(t, h.List, http.MethodGet, "/api/v1/monitoring?projectId=7", "", context.Background())
	// Anonymous requests surface as "project not found" (access lookup bails out).
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}