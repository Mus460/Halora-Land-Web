package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

type AuditLogRepo struct{ pool database.Pool }

func NewAuditLogRepo(pool database.Pool) *AuditLogRepo { return &AuditLogRepo{pool: pool} }

type ListAuditFilter struct {
	UserID     *int32
	ProjectID  *int32
	Action     string
	EntityType string
	Limit      int
	IsAdmin    bool
}

func (r *AuditLogRepo) List(ctx context.Context, f ListAuditFilter) ([]models.AuditLog, error) {
	q := `SELECT id, "projectId", "workItemId", "userId", action, "entityType", "entityId",
		"oldValue", "newValue", description, "ipAddress", "userAgent", "createdAt"
		FROM audit_log WHERE 1=1`
	var args []any
	if !f.IsAdmin && f.UserID != nil {
		args = append(args, *f.UserID)
		q += ` AND "userId" = $` + strconv.Itoa(len(args))
	}
	if f.ProjectID != nil {
		args = append(args, *f.ProjectID)
		q += ` AND "projectId" = $` + strconv.Itoa(len(args))
	}
	if f.Action != "" {
		args = append(args, f.Action)
		q += ` AND action = $` + strconv.Itoa(len(args))
	}
	if f.EntityType != "" {
		args = append(args, f.EntityType)
		q += ` AND "entityType" = $` + strconv.Itoa(len(args))
	}
	q += ` ORDER BY "createdAt" DESC`
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	args = append(args, f.Limit)
	q += ` LIMIT $` + strconv.Itoa(len(args))
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AuditLog
	for rows.Next() {
		var a models.AuditLog
		var projectID, workItemID, entityID sql.NullInt32
		var oldValue, newValue sql.NullString
		var desc, ip, ua sql.NullString
		if err := rows.Scan(&a.ID, &projectID, &workItemID, &a.UserID, &a.Action, &a.EntityType, &entityID,
			&oldValue, &newValue, &desc, &ip, &ua, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.ProjectID = i32Ptr(projectID)
		a.WorkItemID = i32Ptr(workItemID)
		a.EntityID = i32Ptr(entityID)
		if oldValue.Valid && oldValue.String != "" {
			a.OldValue = json.RawMessage(oldValue.String)
		}
		if newValue.Valid && newValue.String != "" {
			a.NewValue = json.RawMessage(newValue.String)
		}
		a.Description = strPtr(desc)
		a.IPAddress = strPtr(ip)
		a.UserAgent = strPtr(ua)
		out = append(out, a)
	}
	return out, rows.Err()
}

// DashboardStats holds the /api/dashboard/stats aggregate (§6.6).
type DashboardStats struct {
	TotalProjects    int32             `json:"totalProjects"`
	ActiveProjects   int32             `json:"activeProjects"`
	PitchingProjects int32             `json:"pitchingProjects"`
	TotalRAB         decimal.Decimal   `json:"totalRAB"`
	TotalWorkItems   int32             `json:"totalWorkItems"`
	RecentProjects   []RecentProject   `json:"recentProjects"`
	RecentAuditLogs  []models.AuditLog `json:"recentAuditLogs"`
}

type RecentProject struct {
	ID        int32           `json:"id"`
	Name      string          `json:"name"`
	Location  *string         `json:"location"`
	TotalRAB  decimal.Decimal `json:"totalRAB"`
	CreatedAt time.Time       `json:"createdAt"`
}

type DashboardRepo struct{ pool database.Pool }

func NewDashboardRepo(pool database.Pool) *DashboardRepo { return &DashboardRepo{pool: pool} }

func (r *DashboardRepo) Stats(ctx context.Context, userID int32, isAdmin bool) (*DashboardStats, error) {
	scope := ``
	var args []any
	if !isAdmin {
		args = append(args, userID)
		scope = ` AND (p."userId" = $1 OR EXISTS (SELECT 1 FROM project_team tp WHERE tp."projectId" = p.id AND tp."userId" = $1))`
	}
	q := `SELECT
		(SELECT count(*) FROM projects p WHERE p."deletedAt" IS NULL` + scope + `),
		(SELECT count(*) FROM projects p WHERE p."deletedAt" IS NULL AND p."isPitching" = false AND p."isDone" = false` + scope + `),
		(SELECT count(*) FROM projects p WHERE p."deletedAt" IS NULL AND p."isPitching" = true` + scope + `),
		(SELECT COALESCE(SUM(pk."totalCost"), 0)::text FROM work_items pk
			JOIN projects p ON p.id = pk."projectId" WHERE pk."deletedAt" IS NULL AND p."deletedAt" IS NULL` + scope + `),
		(SELECT count(*) FROM work_items pk
			JOIN projects p ON p.id = pk."projectId" WHERE pk."deletedAt" IS NULL AND p."deletedAt" IS NULL` + scope + `)`
	var totalProjects, activeProjects, pitchingProjects, totalWorkItems int32
	var rabStr sql.NullString
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&totalProjects, &activeProjects, &pitchingProjects, &rabStr, &totalWorkItems); err != nil {
		return nil, err
	}
	stats := &DashboardStats{
		TotalProjects:    totalProjects,
		ActiveProjects:   activeProjects,
		PitchingProjects: pitchingProjects,
		TotalRAB:         scanDec(rabStr.String),
		TotalWorkItems:   totalWorkItems,
	}

	rq := `SELECT p.id, p."name", p.location, COALESCE(SUM(pk."totalCost"), 0)::text, p."createdAt"
		FROM projects p LEFT JOIN work_items pk ON pk."projectId" = p.id AND pk."deletedAt" IS NULL
		WHERE p."deletedAt" IS NULL` + scope + `
		GROUP BY p.id, p."name", p.location, p."createdAt" ORDER BY p."createdAt" DESC LIMIT 5`
	rrows, err := r.pool.Query(ctx, rq, args...)
	if err != nil {
		return nil, err
	}
	defer rrows.Close()
	for rrows.Next() {
		var rp RecentProject
		var rab sql.NullString
		var location sql.NullString
		if err := rrows.Scan(&rp.ID, &rp.Name, &location, &rab, &rp.CreatedAt); err != nil {
			return nil, err
		}
		rp.TotalRAB = scanDec(rab.String)
		rp.Location = strPtr(location)
		stats.RecentProjects = append(stats.RecentProjects, rp)
	}
	if err := rrows.Err(); err != nil {
		return nil, err
	}

	auditRepo := NewAuditLogRepo(r.pool)
	logs, err := auditRepo.List(ctx, ListAuditFilter{UserID: &userID, IsAdmin: isAdmin, Limit: 10})
	if err != nil {
		return nil, err
	}
	stats.RecentAuditLogs = logs
	return stats, nil
}
