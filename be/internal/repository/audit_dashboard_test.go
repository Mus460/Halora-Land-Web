package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"
)

func TestAuditLogListWithFilters(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM audit_log WHERE`).
		WithArgs(int32(4), int32(1), "CREATE", "work_items", 50).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "workItemId", "userId", "action", "entityType", "entityId",
			"oldValue", "newValue", "description", "ipAddress", "userAgent", "createdAt"}).
			AddRow(int32(1), int32(1), int32(2), int32(4), "CREATE", "work_items", int32(2),
				"{\"a\":1}", nil, "buat item", "1.2.3.4", "curl", time.Now()))
	r := NewAuditLogRepo(m)
	logs, err := r.List(context.Background(), ListAuditFilter{
		UserID: int32Ptr(4), ProjectID: int32Ptr(1), Action: "CREATE", EntityType: "work_items",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs = %d", len(logs))
	}
	a := logs[0]
	if a.ProjectID == nil || *a.ProjectID != 1 {
		t.Errorf("projectId = %v", a.ProjectID)
	}
	if a.WorkItemID == nil || *a.WorkItemID != 2 {
		t.Errorf("workItemId = %v", a.WorkItemID)
	}
	if string(a.OldValue) != `{"a":1}` {
		t.Errorf("oldValue = %s", a.OldValue)
	}
	if len(a.NewValue) != 0 {
		t.Errorf("newValue = %s", a.NewValue)
	}
	if a.Description == nil || *a.Description != "buat item" {
		t.Errorf("desc = %v", a.Description)
	}
	if a.IPAddress == nil || *a.IPAddress != "1.2.3.4" {
		t.Errorf("ip = %v", a.IPAddress)
	}
}

func TestAuditLogListAdminSkipsUserFilter(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM audit_log WHERE`).
		WithArgs(10).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "workItemId", "userId", "action", "entityType", "entityId",
			"oldValue", "newValue", "description", "ipAddress", "userAgent", "createdAt"}).
			AddRow(int32(1), nil, nil, int32(4), "LOGIN", "auth", nil, nil, nil, nil, nil, nil, time.Now()))
	r := NewAuditLogRepo(m)
	logs, err := r.List(context.Background(), ListAuditFilter{IsAdmin: true, Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs = %d", len(logs))
	}
	if logs[0].ProjectID != nil || logs[0].WorkItemID != nil || logs[0].EntityID != nil {
		t.Errorf("nulls = %+v", logs[0])
	}
}

func TestAuditLogListLimitClamp(t *testing.T) {
	m := newPool(t)
	// limit out of range → clamped to 50
	m.ExpectQuery(`FROM audit_log WHERE`).
		WithArgs(int32(4), 50).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "workItemId", "userId", "action", "entityType", "entityId",
			"oldValue", "newValue", "description", "ipAddress", "userAgent", "createdAt"}))
	r := NewAuditLogRepo(m)
	_, err := r.List(context.Background(), ListAuditFilter{UserID: int32Ptr(4), Limit: 9999})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestDashboardStatsAdmin(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM projects p WHERE p.`).
		WillReturnRows(pgxmock.NewRows([]string{"totalProjects", "activeProjects", "pitchingProjects", "totalRAB", "totalWorkItems"}).
			AddRow(int32(5), int32(3), int32(1), "1500000000", int32(42)))
	m.ExpectQuery(`GROUP BY p.id`).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "location", "totalRAB", "createdAt"}).
			AddRow(int32(1), "Ruko", "Jakarta", "500000000", time.Now()))
	m.ExpectQuery(`FROM audit_log WHERE`).
		WithArgs(10).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "workItemId", "userId", "action", "entityType", "entityId",
			"oldValue", "newValue", "description", "ipAddress", "userAgent", "createdAt"}).
			AddRow(int32(9), nil, nil, int32(4), "LOGIN", "auth", nil, nil, nil, nil, nil, nil, time.Now()))

	r := NewDashboardRepo(m)
	stats, err := r.Stats(context.Background(), 4, true)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalProjects != 5 || stats.ActiveProjects != 3 || stats.PitchingProjects != 1 {
		t.Errorf("stats = %+v", stats)
	}
	if !stats.TotalRAB.Equal(decimalNew("1500000000")) {
		t.Errorf("totalRAB = %s", stats.TotalRAB)
	}
	if stats.TotalWorkItems != 42 {
		t.Errorf("totalWorkItems = %d", stats.TotalWorkItems)
	}
	if len(stats.RecentProjects) != 1 || stats.RecentProjects[0].Name != "Ruko" {
		t.Errorf("recent = %+v", stats.RecentProjects)
	}
	if len(stats.RecentAuditLogs) != 1 {
		t.Errorf("audit = %+v", stats.RecentAuditLogs)
	}
}

func TestDashboardStatsScopedByUser(t *testing.T) {
	m := newPool(t)
	m.ExpectQuery(`FROM projects p WHERE p.`).WithArgs(int32(4)).
		WillReturnRows(pgxmock.NewRows([]string{"totalProjects", "activeProjects", "pitchingProjects", "totalRAB", "totalWorkItems"}).
			AddRow(int32(2), int32(1), int32(0), "0", int32(3)))
	m.ExpectQuery(`GROUP BY p.id`).WithArgs(int32(4)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "location", "totalRAB", "createdAt"}))
	m.ExpectQuery(`FROM audit_log WHERE`).
		WithArgs(int32(4), 10).
		WillReturnRows(pgxmock.NewRows([]string{"id", "projectId", "workItemId", "userId", "action", "entityType", "entityId",
			"oldValue", "newValue", "description", "ipAddress", "userAgent", "createdAt"}))

	r := NewDashboardRepo(m)
	stats, err := r.Stats(context.Background(), 4, false)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalProjects != 2 {
		t.Errorf("stats = %+v", stats)
	}
}

func decimalNew(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}
