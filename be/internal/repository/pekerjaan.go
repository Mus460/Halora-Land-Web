package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

type WorkItemRepo struct{ pool database.Pool }

var ErrProgressNotIncreasing = errors.New("progress baru harus lebih tinggi dari progress saat ini")

func NewWorkItemRepo(pool database.Pool) *WorkItemRepo { return &WorkItemRepo{pool: pool} }

type ListWorkItemFilter struct {
	ProjectID *int32
	Category  string
	Search    string
}

func (r *WorkItemRepo) List(ctx context.Context, f ListWorkItemFilter) ([]models.WorkItem, error) {
	q := `SELECT id, "projectId", category, "description", volume, unit, "unitPrice", "totalCost",
		"calculationMethod", "level", "type", "analysisMasterId", "basePrice", duration, (duration * volume) AS "totalDuration",
		"createdAt", "updatedAt"
		FROM work_items WHERE "deletedAt" IS NULL`
	var args []any
	if f.ProjectID != nil {
		args = append(args, *f.ProjectID)
		q += ` AND "projectId" = $1`
	}
	if f.Category != "" {
		args = append(args, f.Category)
		q += ` AND category = $` + strconv.Itoa(len(args))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += ` AND "description" ILIKE $` + strconv.Itoa(len(args))
	}
	q += ` ORDER BY id ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.WorkItem
	for rows.Next() {
		p, err := scanWorkItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func argPlaceholder(i int) string { return strconv.Itoa(i) }

func scanWorkItem(s rowScanner) (*models.WorkItem, error) {
	var p models.WorkItem
	var vol, hs, tb string
	var level, wtype, basePrice, duration, totalDuration sql.NullString
	var maID sql.NullInt32
	if err := s.Scan(&p.ID, &p.ProjectID, &p.Category, &p.Description, &vol, &p.Unit, &hs, &tb,
		&p.CalculationMethod, &level, &wtype, &maID, &basePrice, &duration, &totalDuration, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Volume = scanDec(vol)
	p.UnitPrice = scanDec(hs)
	p.TotalCost = scanDec(tb)
	p.Level = strPtr(level)
	p.Type = strPtr(wtype)
	p.AnalysisMasterID = i32Ptr(maID)
	p.BasePrice = scanDecPtr(basePrice)
	p.Duration = scanDecPtr(duration)
	p.TotalDuration = scanDecPtr(totalDuration)
	return &p, nil
}

func (r *WorkItemRepo) Get(ctx context.Context, id int32) (*models.WorkItem, error) {
	p, err := scanWorkItem(r.pool.QueryRow(ctx, `
		SELECT id, "projectId", category, "description", volume, unit, "unitPrice", "totalCost",
		"calculationMethod", "level", "type", "analysisMasterId", "basePrice", duration, (duration * volume) AS "totalDuration",
		"createdAt", "updatedAt"
		FROM work_items WHERE id = $1 AND "deletedAt" IS NULL`, id))
	if err != nil {
		return nil, err
	}
	details, err := r.ListWorkItemDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	p.ItemDetails = details
	return p, nil
}

func (r *WorkItemRepo) GetByID(ctx context.Context, id int32) (*models.WorkItem, error) {
	p, err := scanWorkItem(r.pool.QueryRow(ctx, `
		SELECT id, "projectId", category, "description", volume, unit, "unitPrice", "totalCost",
		"calculationMethod", "level", "type", "analysisMasterId", "basePrice", duration, (duration * volume) AS "totalDuration",
		"createdAt", "updatedAt"
		FROM work_items WHERE id = $1 AND "deletedAt" IS NULL`, id))
	if err != nil {
		return nil, err
	}
	return p, nil
}

// DurationCoefficient returns the time coefficient (jam per unit) for a master analisa
// item, computed as the sum of duration over its rincian rows. Nil when absent.
func (r *WorkItemRepo) DurationCoefficient(ctx context.Context, analysisMasterID int32) (*decimal.Decimal, error) {
	var s sql.NullString
	err := r.pool.QueryRow(ctx, `
		SELECT SUM(duration) FROM analysis_components WHERE "analysisMasterId" = $1 AND "deletedAt" IS NULL`, analysisMasterID).Scan(&s)
	if err != nil {
		return nil, err
	}
	return scanDecPtr(s), nil
}

type CreateWorkItemInput struct {
	ProjectID         int32
	Category          models.WorkCategory
	Description       string
	Volume            decimal.Decimal
	Unit              string
	UnitPrice         decimal.Decimal
	TotalCost         decimal.Decimal
	CalculationMethod models.CalculationMethod
	Level             *string
	Type              *string
	AnalysisMasterID  *int32
	BasePrice         *decimal.Decimal
	Duration          *decimal.Decimal
}

func (r *WorkItemRepo) Create(ctx context.Context, tx pgx.Tx, in CreateWorkItemInput) (*models.WorkItem, error) {
	exec := r.execer(tx)
	row := exec.QueryRow(ctx, `
		INSERT INTO work_items ("projectId", category, "description", volume, unit, "unitPrice", "totalCost",
			"calculationMethod", "level", "type", "analysisMasterId", "basePrice", duration)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, "projectId", category, "description", volume, unit, "unitPrice", "totalCost",
			"calculationMethod", "level", "type", "analysisMasterId", "basePrice", duration, (duration * volume) AS "totalDuration",
			"createdAt", "updatedAt"`,
		in.ProjectID, in.Category, in.Description, decArg(in.Volume), in.Unit, decArg(in.UnitPrice), decArg(in.TotalCost),
		in.CalculationMethod, in.Level, in.Type, in.AnalysisMasterID, decPtrArg(in.BasePrice), decPtrArg(in.Duration))
	return scanWorkItem(row)
}

type UpdateWorkItemInput struct {
	Volume            *decimal.Decimal
	UnitPrice         *decimal.Decimal
	TotalCost         *decimal.Decimal
	Description       *string
	Unit              *string
	Level             *string
	Type              *string
	CalculationMethod *models.CalculationMethod
}

func (r *WorkItemRepo) Update(ctx context.Context, id int32, in UpdateWorkItemInput) (*models.WorkItem, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE work_items SET
			volume = COALESCE($2, volume),
			"unitPrice" = COALESCE($3, "unitPrice"),
			"totalCost" = COALESCE($4, "totalCost"),
			"description" = COALESCE($5, "description"),
			unit = COALESCE($6, unit),
			"level" = COALESCE($7, "level"),
			"type" = COALESCE($8, "type"),
			"calculationMethod" = COALESCE($9, "calculationMethod"),
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, "projectId", category, "description", volume, unit, "unitPrice", "totalCost",
			"calculationMethod", "level", "type", "analysisMasterId", "basePrice", duration, (duration * volume) AS "totalDuration",
			"createdAt", "updatedAt"`,
		id, decPtrArg(in.Volume), decPtrArg(in.UnitPrice), decPtrArg(in.TotalCost),
		in.Description, in.Unit, in.Level, in.Type, in.CalculationMethod)
	return scanWorkItem(row)
}

func (r *WorkItemRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `UPDATE work_items SET "deletedAt" = NOW() WHERE id = $1`, id)
	return err
}

func (r *WorkItemRepo) SetProgress(ctx context.Context, id int32, progress int) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	_, err := r.pool.Exec(ctx, `UPDATE work_items SET progress = $2, "updatedAt" = NOW() WHERE id = $1 AND "deletedAt" IS NULL`, id, progress)
	return err
}

// RecordProgress updates the current progress value and appends a timestamped
// log entry when the value actually changed. Returns the created log.
func (r *WorkItemRepo) RecordProgress(ctx context.Context, id int32, progress int, note *string) (*models.WorkItemProgressLog, error) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	var current int
	if err := tx.QueryRow(ctx, `SELECT progress FROM work_items WHERE id = $1 AND "deletedAt" IS NULL`, id).Scan(&current); err != nil {
		return nil, err
	}
	if progress <= current {
		return nil, ErrProgressNotIncreasing
	}
	if _, err := tx.Exec(ctx, `UPDATE work_items SET progress = $2, "updatedAt" = NOW() WHERE id = $1`, id, progress); err != nil {
		return nil, err
	}
	if progress == current {
		if err := tx.Commit(context.Background()); err != nil {
			return nil, err
		}
		return nil, nil
	}
	log := &models.WorkItemProgressLog{WorkItemID: id, Progress: progress, Note: note}
	if err := tx.QueryRow(ctx, `
		INSERT INTO work_item_progress_logs ("workItemId", progress, note)
		VALUES ($1,$2,$3) RETURNING id, "createdAt"`, id, progress, note).Scan(&log.ID, &log.CreatedAt); err != nil {
		return nil, err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}
	return log, nil
}

func (r *WorkItemRepo) ListProgressLogs(ctx context.Context, workItemID int32) ([]models.WorkItemProgressLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "workItemId", progress, note, "createdAt"
		FROM work_item_progress_logs WHERE "workItemId" = $1 ORDER BY "createdAt" ASC, id ASC`, workItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.WorkItemProgressLog
	for rows.Next() {
		var l models.WorkItemProgressLog
		var note sql.NullString
		if err := rows.Scan(&l.ID, &l.WorkItemID, &l.Progress, &note, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.Note = strPtr(note)
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *WorkItemRepo) execer(tx pgx.Tx) executor {
	if tx != nil {
		return tx
	}
	return r.pool
}

type executor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type CreateDetailInput struct {
	WorkItemID       int32
	PriceMasterID    *int32
	AnalysisMasterID *int32
	Name             string
	Unit             string
	Coefficient      decimal.Decimal
	UnitPrice        decimal.Decimal
	TotalCost        decimal.Decimal
	Type             models.ComponentType
	SourceCode       *string
}

func (r *WorkItemRepo) CreateDetail(ctx context.Context, tx pgx.Tx, in CreateDetailInput) error {
	exec := r.execer(tx)
	_, err := exec.Exec(ctx, `
		INSERT INTO work_item_details ("workItemId", "priceMasterId", "analysisMasterId", name, unit, coefficient,
			"unitPrice", "totalCost", type, "snapshotAt", "sourceCode")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),$10)`,
		in.WorkItemID, in.PriceMasterID, in.AnalysisMasterID, in.Name, in.Unit, decArg(in.Coefficient),
		decArg(in.UnitPrice), decArg(in.TotalCost), in.Type, in.SourceCode)
	return err
}

func (r *WorkItemRepo) ListWorkItemDetail(ctx context.Context, workItemID int32) ([]models.WorkItemDetail, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "workItemId", "priceMasterId", "analysisMasterId", name, unit, coefficient, "unitPrice",
			"totalCost", type, "snapshotAt", "sourceCode"
		FROM work_item_details WHERE "workItemId" = $1 ORDER BY id ASC`, workItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.WorkItemDetail
	for rows.Next() {
		d, err := scanDetail(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func scanDetail(s rowScanner) (*models.WorkItemDetail, error) {
	var d models.WorkItemDetail
	var mhID, maID sql.NullInt32
	var coefficient, hs, tb string
	var src sql.NullString
	if err := s.Scan(&d.ID, &d.WorkItemID, &mhID, &maID, &d.Name, &d.Unit, &coefficient, &hs, &tb, &d.Type, &d.SnapshotAt, &src); err != nil {
		return nil, err
	}
	d.PriceMasterID = i32Ptr(mhID)
	d.AnalysisMasterID = i32Ptr(maID)
	d.Coefficient = scanDec(coefficient)
	d.UnitPrice = scanDec(hs)
	d.TotalCost = scanDec(tb)
	d.SourceCode = strPtr(src)
	return &d, nil
}

// DeleteDetails removes all frozen breakdown rows for a work_items (used by recalculate).
func (r *WorkItemRepo) DeleteDetails(ctx context.Context, tx pgx.Tx, workItemID int32) error {
	exec := r.execer(tx)
	_, err := exec.Exec(ctx, `DELETE FROM work_item_details WHERE "workItemId" = $1`, workItemID)
	return err
}

type UpdateDetailInput struct {
	ID          int32
	WorkItemID  int32
	Name        string
	Unit        string
	Coefficient decimal.Decimal
	UnitPrice   decimal.Decimal
	TotalCost   decimal.Decimal
	Type        models.ComponentType
}

// UpdateDetail edits an existing local breakdown row in place (keeps its
// sourceCode/priceMasterId/analysisMasterId lineage).
func (r *WorkItemRepo) UpdateDetail(ctx context.Context, tx pgx.Tx, in UpdateDetailInput) error {
	exec := r.execer(tx)
	_, err := exec.Exec(ctx, `
		UPDATE work_item_details SET
			name = $3, unit = $4, coefficient = $5, "unitPrice" = $6, "totalCost" = $7, type = $8
		WHERE id = $1 AND "workItemId" = $2`,
		in.ID, in.WorkItemID, in.Name, in.Unit, decArg(in.Coefficient), decArg(in.UnitPrice), decArg(in.TotalCost), in.Type)
	return err
}

// DeleteDetailsNotIn removes local breakdown rows not present in keepIDs
// (bulk-replace semantics). Empty keepIDs clears the whole set.
func (r *WorkItemRepo) DeleteDetailsNotIn(ctx context.Context, tx pgx.Tx, workItemID int32, keepIDs []int32) error {
	exec := r.execer(tx)
	if len(keepIDs) == 0 {
		_, err := exec.Exec(ctx, `DELETE FROM work_item_details WHERE "workItemId" = $1`, workItemID)
		return err
	}
	_, err := exec.Exec(ctx, `DELETE FROM work_item_details WHERE "workItemId" = $1 AND NOT (id = ANY($2))`, workItemID, keepIDs)
	return err
}

// SetBasePrice stores the master reference price (used by recalculate).
func (r *WorkItemRepo) SetBasePrice(ctx context.Context, tx pgx.Tx, id int32, base *decimal.Decimal) error {
	exec := r.execer(tx)
	_, err := exec.Exec(ctx, `UPDATE work_items SET "basePrice" = $2, "updatedAt" = NOW() WHERE id = $1`, id, decPtrArg(base))
	return err
}

func (r *WorkItemRepo) SetTotal(ctx context.Context, tx pgx.Tx, id int32, hs, tb decimal.Decimal) error {
	exec := r.execer(tx)
	_, err := exec.Exec(ctx, `UPDATE work_items SET "unitPrice" = $2, "totalCost" = $3, "updatedAt" = NOW() WHERE id = $1`,
		id, decArg(hs), decArg(tb))
	return err
}
