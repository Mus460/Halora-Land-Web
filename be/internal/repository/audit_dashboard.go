package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

type AuditLogRepo struct{ pool *pgxpool.Pool }

func NewAuditLogRepo(pool *pgxpool.Pool) *AuditLogRepo { return &AuditLogRepo{pool: pool} }

type ListAuditFilter struct {
	UserID     *int32
	ProyekID   *int32
	Action     string
	EntityType string
	Limit      int
	IsAdmin    bool
}

func (r *AuditLogRepo) List(ctx context.Context, f ListAuditFilter) ([]models.AuditLog, error) {
	q := `SELECT id, "proyekId", "pekerjaanId", "userId", action, "entityType", "entityId",
		"oldValue", "newValue", description, "ipAddress", "userAgent", "createdAt"
		FROM audit_log WHERE 1=1`
	var args []any
	if !f.IsAdmin && f.UserID != nil {
		args = append(args, *f.UserID)
		q += ` AND "userId" = $` + strconv.Itoa(len(args))
	}
	if f.ProyekID != nil {
		args = append(args, *f.ProyekID)
		q += ` AND "proyekId" = $` + strconv.Itoa(len(args))
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
		var proyekID, pekerjaanID, entityID sql.NullInt32
		var oldValue, newValue sql.NullString
		var desc, ip, ua sql.NullString
		if err := rows.Scan(&a.ID, &proyekID, &pekerjaanID, &a.UserID, &a.Action, &a.EntityType, &entityID,
			&oldValue, &newValue, &desc, &ip, &ua, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.ProyekID = i32Ptr(proyekID)
		a.PekerjaanID = i32Ptr(pekerjaanID)
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
	TotalProyek      int32           `json:"totalProyek"`
	ProyekAktif      int32           `json:"proyekAktif"`
	ProyekPitching   int32           `json:"proyekPitching"`
	TotalRAB         decimal.Decimal `json:"totalRAB"`
	TotalPekerjaan   int32           `json:"totalPekerjaan"`
	RecentProjects   []RecentProject `json:"recentProjects"`
	RecentAuditLogs  []models.AuditLog `json:"recentAuditLogs"`
}

type RecentProject struct {
	ID         int32            `json:"id"`
	Nama       string           `json:"nama"`
	Lokasi     *string          `json:"lokasi"`
	TotalRAB   decimal.Decimal  `json:"totalRAB"`
	CreatedAt  time.Time        `json:"createdAt"`
}

type DashboardRepo struct{ pool *pgxpool.Pool }

func NewDashboardRepo(pool *pgxpool.Pool) *DashboardRepo { return &DashboardRepo{pool: pool} }

func (r *DashboardRepo) Stats(ctx context.Context, userID int32, isAdmin bool) (*DashboardStats, error) {
	scope := ``
	var args []any
	if !isAdmin {
		args = append(args, userID)
		scope = ` AND (p."userId" = $1 OR EXISTS (SELECT 1 FROM tim_proyek tp WHERE tp."proyekId" = p.id AND tp."userId" = $1))`
	}
	q := `SELECT
		(SELECT count(*) FROM proyek p WHERE p."deletedAt" IS NULL` + scope + `),
		(SELECT count(*) FROM proyek p WHERE p."deletedAt" IS NULL AND p."isPitching" = false` + scope + `),
		(SELECT count(*) FROM proyek p WHERE p."deletedAt" IS NULL AND p."isPitching" = true` + scope + `),
		(SELECT COALESCE(SUM(pk."totalBiaya"), 0)::text FROM pekerjaan pk
			JOIN proyek p ON p.id = pk."proyekId" WHERE pk."deletedAt" IS NULL AND p."deletedAt" IS NULL` + scope + `),
		(SELECT count(*) FROM pekerjaan pk
			JOIN proyek p ON p.id = pk."proyekId" WHERE pk."deletedAt" IS NULL AND p."deletedAt" IS NULL` + scope + `)`
	var totalProyek, proyekAktif, proyekPitching, totalPekerjaan int32
	var rabStr sql.NullString
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&totalProyek, &proyekAktif, &proyekPitching, &rabStr, &totalPekerjaan); err != nil {
		return nil, err
	}
	stats := &DashboardStats{
		TotalProyek:    totalProyek,
		ProyekAktif:    proyekAktif,
		ProyekPitching: proyekPitching,
		TotalRAB:       scanDec(rabStr.String),
		TotalPekerjaan: totalPekerjaan,
	}

	rq := `SELECT p.id, p."namaProyek", p.lokasi, COALESCE(SUM(pk."totalBiaya"), 0)::text, p."createdAt"
		FROM proyek p LEFT JOIN pekerjaan pk ON pk."proyekId" = p.id AND pk."deletedAt" IS NULL
		WHERE p."deletedAt" IS NULL` + scope + `
		GROUP BY p.id, p."namaProyek", p.lokasi, p."createdAt" ORDER BY p."createdAt" DESC LIMIT 5`
	rrows, err := r.pool.Query(ctx, rq, args...)
	if err != nil {
		return nil, err
	}
	defer rrows.Close()
	for rrows.Next() {
		var rp RecentProject
		var rab sql.NullString
		var lokasi sql.NullString
		if err := rrows.Scan(&rp.ID, &rp.Nama, &lokasi, &rab, &rp.CreatedAt); err != nil {
			return nil, err
		}
		rp.TotalRAB = scanDec(rab.String)
		rp.Lokasi = strPtr(lokasi)
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
