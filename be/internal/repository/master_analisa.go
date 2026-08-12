package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/cache"
	"github.com/halora-land/halora-be/internal/models"
)

type AnalysisMasterRepo struct {
	pool  database.Pool
	cache *cache.Cache
}

func NewAnalysisMasterRepo(pool database.Pool) *AnalysisMasterRepo {
	return &AnalysisMasterRepo{pool: pool, cache: cache.New(60 * time.Second)}
}

type ListAnalysisMasterFilter struct {
	Level    *int
	ParentID *int32
	Search   string
	IsGlobal *bool
	UserID   int32
}

func (f ListAnalysisMasterFilter) cacheKey() string {
	level, parentID, isGlobal := "nil", "nil", "nil"
	if f.Level != nil {
		level = strconv.Itoa(*f.Level)
	}
	if f.ParentID != nil {
		parentID = strconv.FormatInt(int64(*f.ParentID), 10)
	}
	if f.IsGlobal != nil {
		isGlobal = strconv.FormatBool(*f.IsGlobal)
	}
	return fmt.Sprintf("analisa|u:%d|level:%s|parent:%s|search:%s|global:%s", f.UserID, level, parentID, f.Search, isGlobal)
}

func (r *AnalysisMasterRepo) List(ctx context.Context, f ListAnalysisMasterFilter) ([]models.AnalysisMaster, error) {
	if v, ok := r.cache.Get(f.cacheKey()); ok {
		return v.([]models.AnalysisMaster), nil
	}
	q := `SELECT id, code, name, level, "parentId", unit, "unitPrice", category, "isGlobal", "userId",
		"isSystem", "ahspCode", "ahspSheet", "generalCost", "createdAt", "updatedAt"
		FROM analysis_masters WHERE ("userId" = $1 OR "isGlobal" = true OR "isSystem" = true) AND "deletedAt" IS NULL`
	args := []any{f.UserID}
	if f.Level != nil {
		args = append(args, *f.Level)
		q += ` AND level = $` + strconv.Itoa(len(args))
	}
	if f.ParentID != nil {
		args = append(args, *f.ParentID)
		q += ` AND "parentId" = $` + strconv.Itoa(len(args))
	} else if f.Level != nil {
		args = append(args, nil)
		q += ` AND "parentId" IS NOT DISTINCT FROM $` + strconv.Itoa(len(args))
	}
	if f.IsGlobal != nil {
		args = append(args, *f.IsGlobal)
		q += ` AND "isGlobal" = $` + strconv.Itoa(len(args))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += ` AND (name ILIKE $` + strconv.Itoa(len(args))
		args = append(args, "%"+f.Search+"%")
		q += ` OR "ahspCode" ILIKE $` + strconv.Itoa(len(args)) + `)`
	}
	q += ` ORDER BY level ASC, code ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AnalysisMaster
	for rows.Next() {
		m, err := scanAnalysisMaster(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	r.cache.Set(f.cacheKey(), out)
	return out, nil
}

func scanAnalysisMaster(s rowScanner) (*models.AnalysisMaster, error) {
	var m models.AnalysisMaster
	var parentID sql.NullInt32
	var unit, category, ahspCode, ahspSheet, hs, generalCost sql.NullString
	var userID sql.NullInt32
	if err := s.Scan(&m.ID, &m.Code, &m.Name, &m.Level, &parentID, &unit, &hs, &category,
		&m.IsGlobal, &userID, &m.IsSystem, &ahspCode, &ahspSheet, &generalCost, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	m.ParentID = i32Ptr(parentID)
	m.Unit = strPtr(unit)
	m.UnitPrice = scanDecPtr(hs)
	m.Category = strPtr(category)
	m.UserID = i32Ptr(userID)
	m.AHSPCode = strPtr(ahspCode)
	m.AHSPSheet = strPtr(ahspSheet)
	m.GeneralCost = scanDec(generalCost.String)
	return &m, nil
}

func (r *AnalysisMasterRepo) Get(ctx context.Context, id int32) (*models.AnalysisMaster, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, code, name, level, "parentId", unit, "unitPrice", category, "isGlobal", "userId",
		"isSystem", "ahspCode", "ahspSheet", "generalCost", "createdAt", "updatedAt"
		FROM analysis_masters WHERE id = $1 AND "deletedAt" IS NULL`, id)
	return scanAnalysisMaster(row)
}

type CreateAnalysisMasterInput struct {
	Code     string
	Name     string
	Level    int32
	ParentID *int32
	Unit     *string
	IsGlobal bool
	UserID   *int32
	IsSystem bool
}

func (r *AnalysisMasterRepo) Create(ctx context.Context, in CreateAnalysisMasterInput) (*models.AnalysisMaster, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO analysis_masters (code, name, level, "parentId", unit, "isGlobal", "userId", "isSystem")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, code, name, level, "parentId", unit, "unitPrice", category, "isGlobal", "userId",
		"isSystem", "ahspCode", "ahspSheet", "generalCost", "createdAt", "updatedAt"`,
		in.Code, in.Name, in.Level, in.ParentID, in.Unit, in.IsGlobal, in.UserID, in.IsSystem)
	m, err := scanAnalysisMaster(row)
	if err != nil {
		return nil, err
	}
	r.cache.Clear()
	return m, nil
}

func (r *AnalysisMasterRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `UPDATE analysis_masters SET "deletedAt" = NOW() WHERE id = $1`, id)
	if err == nil {
		r.cache.Clear()
	}
	return err
}

// Copy duplicates an analysis master (and its components, plus any children)
// into a user-owned, editable row that is detached from AHSP source data.
func (r *AnalysisMasterRepo) Copy(ctx context.Context, id, userID int32, newName string) (*models.AnalysisMaster, error) {
	src, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	name := newName
	if name == "" {
		name = "Salin - " + src.Name
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	newCode := func(srcCode string) (string, error) {
		candidate := srcCode
		suffix := 2
		for {
			var exists bool
			err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM analysis_masters WHERE code = $1 AND "userId" = $2 AND "deletedAt" IS NULL)`, candidate, userID).Scan(&exists)
			if err != nil {
				return "", err
			}
			if !exists {
				return candidate, nil
			}
			candidate = fmt.Sprintf("%s-%d", srcCode, suffix)
			suffix++
		}
	}

	var copyNode func(srcID, newParentID int32) (int32, error)
	copyNode = func(srcID, newParentID int32) (int32, error) {
		var srcChild models.AnalysisMaster
		m, err := scanAnalysisMaster(tx.QueryRow(ctx, `
			SELECT id, code, name, level, "parentId", unit, "unitPrice", category, "isGlobal", "userId",
			"isSystem", "ahspCode", "ahspSheet", "generalCost", "createdAt", "updatedAt"
			FROM analysis_masters WHERE id = $1 AND "deletedAt" IS NULL`, srcID))
		if err != nil {
			return 0, err
		}
		srcChild = *m
		code, err := newCode(srcChild.Code)
		if err != nil {
			return 0, err
		}
		var newID int32
		if err := tx.QueryRow(ctx, `
			INSERT INTO analysis_masters (code, name, level, "parentId", unit, "unitPrice", category, "isGlobal", "userId", "isSystem", "generalCost")
			VALUES ($1,$2,$3,NULLIF($4,0),$5,$6,$7,false,$8,false,$9)
			RETURNING id`,
			code, name, srcChild.Level, newParentID, srcChild.Unit, decPtrArg(srcChild.UnitPrice),
			srcChild.Category, userID, decArg(srcChild.GeneralCost)).Scan(&newID); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO analysis_components ("analysisMasterId", "componentId", coefficient, type, name, unit, "unitPrice", "totalPrice", "referenceCode", duration, sequence)
			SELECT $1, "componentId", coefficient, type, name, unit, "unitPrice", "totalPrice", "referenceCode", duration, sequence
			FROM analysis_components WHERE "analysisMasterId" = $2 AND "deletedAt" IS NULL`, newID, srcID); err != nil {
			return 0, err
		}
		rows, err := tx.Query(ctx, `SELECT id FROM analysis_masters WHERE "parentId" = $1 AND "deletedAt" IS NULL ORDER BY code`, srcID)
		if err != nil {
			return 0, err
		}
		var childIDs []int32
		for rows.Next() {
			var c int32
			if err := rows.Scan(&c); err != nil {
				rows.Close()
				return 0, err
			}
			childIDs = append(childIDs, c)
		}
		rows.Close()
		for _, c := range childIDs {
			if _, err := copyNode(c, newID); err != nil {
				return 0, err
			}
		}
		return newID, nil
	}

	newID, err := copyNode(id, 0)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}
	r.cache.Clear()
	return r.Get(ctx, newID)
}

type UpdateAnalysisMasterInput struct {
	Name *string
	Unit *string
}

func (r *AnalysisMasterRepo) Update(ctx context.Context, id, userID int32, isAdmin bool, in UpdateAnalysisMasterInput) (*models.AnalysisMaster, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE analysis_masters SET name = COALESCE($1, name), unit = COALESCE($2, unit), "updatedAt" = NOW()
		WHERE id = $3 AND "deletedAt" IS NULL AND "isSystem" = false AND ("userId" = $4 OR $5)
		RETURNING id, code, name, level, "parentId", unit, "unitPrice", category, "isGlobal", "userId",
		"isSystem", "ahspCode", "ahspSheet", "generalCost", "createdAt", "updatedAt"`,
		in.Name, in.Unit, id, userID, isAdmin)
	m, err := scanAnalysisMaster(row)
	if err == nil {
		r.cache.Clear()
	}
	return m, err
}

func (r *AnalysisMasterRepo) HasChildren(ctx context.Context, id int32) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM analysis_masters WHERE "parentId" = $1 AND "deletedAt" IS NULL`, id).Scan(&n)
	return n > 0, err
}

// --- analysis_components (template breakdown, §3.2) ---

func (r *AnalysisMasterRepo) ListComponents(ctx context.Context, analysisMasterID int32) ([]models.AnalysisComponent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "analysisMasterId", "componentId", coefficient, type, name, unit, "unitPrice", "totalPrice",
		"referenceCode", duration, sequence, "createdAt", "updatedAt"
		FROM analysis_components WHERE "analysisMasterId" = $1 AND "deletedAt" IS NULL ORDER BY sequence ASC, id ASC`, analysisMasterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AnalysisComponent
	for rows.Next() {
		rin, err := scanAnalysisComponent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rin)
	}
	return out, rows.Err()
}

func scanAnalysisComponent(s rowScanner) (*models.AnalysisComponent, error) {
	var rin models.AnalysisComponent
	var komponenID sql.NullInt32
	var coefficient string
	var name, unit, hs, jh, kodeRef, duration sql.NullString
	if err := s.Scan(&rin.ID, &rin.AnalysisMasterID, &komponenID, &coefficient, &rin.Type, &name, &unit, &hs, &jh, &kodeRef, &duration, &rin.Sequence, &rin.CreatedAt, &rin.UpdatedAt); err != nil {
		return nil, err
	}
	rin.ComponentID = i32Ptr(komponenID)
	rin.Coefficient = scanDec(coefficient)
	rin.Name = strPtr(name)
	rin.Unit = strPtr(unit)
	rin.UnitPrice = scanDecPtr(hs)
	rin.TotalPrice = scanDecPtr(jh)
	rin.ReferenceCode = strPtr(kodeRef)
	rin.Duration = scanDecPtr(duration)
	return &rin, nil
}

type UpdateComponentInput struct {
	ID               int32
	AnalysisMasterID int32
	Coefficient      *decimal.Decimal
	UnitPrice        *decimal.Decimal
	Unit             *string
	Name             *string
	Type             *models.ComponentType
}

// UpdateComponent edits a breakdown row of a non-system, user-owned master,
// then recomputes the component totalPrice and the master unitPrice.
func (r *AnalysisMasterRepo) UpdateComponent(ctx context.Context, userID int32, isAdmin bool, in UpdateComponentInput) (*models.AnalysisComponent, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE analysis_components c SET
			coefficient = COALESCE($1, c.coefficient),
			"unitPrice" = COALESCE($2, c."unitPrice"),
			unit = COALESCE($3, c.unit),
			name = COALESCE($4, c.name),
			type = COALESCE($5, c.type),
			"totalPrice" = CASE
				WHEN (COALESCE($2, c."unitPrice") IS NOT NULL)
				THEN ROUND(COALESCE($1, c.coefficient) * COALESCE($2, c."unitPrice") * 100) / 100
				ELSE c."totalPrice" END,
			"updatedAt" = NOW()
		FROM analysis_masters m
		WHERE c.id = $6 AND c."analysisMasterId" = m.id AND c."deletedAt" IS NULL
			AND m."deletedAt" IS NULL AND m."isSystem" = false AND (m."userId" = $7 OR $8)
		RETURNING c.id, c."analysisMasterId", c."componentId", c.coefficient, c.type, c.name, c.unit, c."unitPrice", c."totalPrice",
		"referenceCode", duration, sequence, c."createdAt", c."updatedAt"`,
		decPtrArg(in.Coefficient), decPtrArg(in.UnitPrice), in.Unit, in.Name, in.Type,
		in.ID, userID, isAdmin)
	comp, err := scanAnalysisComponent(row)
	if err != nil {
		return nil, err
	}
	if _, err := r.pool.Exec(ctx, `
		UPDATE analysis_masters m SET
			"unitPrice" = ROUND((SELECT COALESCE(SUM(COALESCE(c."totalPrice",0)),0) FROM analysis_components c
				WHERE c."analysisMasterId" = m.id AND c."deletedAt" IS NULL) * (1 + COALESCE(m."generalCost",0)), 2),
			"updatedAt" = NOW()
		WHERE m.id = $1 AND m."deletedAt" IS NULL`, in.AnalysisMasterID); err != nil {
		return nil, err
	}
	r.cache.Clear()
	return comp, nil
}

type CreateComponentInput struct {
	AnalysisMasterID int32                `json:"analysisMasterId"`
	ComponentID      *int32               `json:"componentId"`
	Coefficient      decimal.Decimal      `json:"coefficient"`
	Type             models.ComponentType `json:"type"`
}

func (r *AnalysisMasterRepo) CreateComponent(ctx context.Context, in CreateComponentInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO analysis_components ("analysisMasterId", "componentId", coefficient, type)
		VALUES ($1,$2,$3,$4)`, in.AnalysisMasterID, in.ComponentID, decArg(in.Coefficient), in.Type)
	return err
}

func (r *AnalysisMasterRepo) DeleteComponent(ctx context.Context, analysisMasterID, componentID int32) error {
	_, err := r.pool.Exec(ctx, `UPDATE analysis_components SET "deletedAt" = NOW() WHERE id = $1 AND "analysisMasterId" = $2`, componentID, analysisMasterID)
	return err
}

// ListTree fetches all accessible nodes and builds a tree from the root level.
func (r *AnalysisMasterRepo) ListTree(ctx context.Context, f ListAnalysisMasterFilter) ([]models.AnalysisMaster, error) {
	treeFilter := f
	treeFilter.Level = nil
	treeFilter.ParentID = nil
	all, err := r.List(ctx, treeFilter)
	if err != nil {
		return nil, err
	}
	return buildAnalysisMasterTree(all), nil
}

// buildAnalysisMasterTree assembles the parent/child tree. all is sorted by
// level ASC (parents before children), so we wire children deepest-first:
// every node is fully assembled before it is copied into its parent, and root
// copies are taken only after the whole tree is wired.
func buildAnalysisMasterTree(all []models.AnalysisMaster) []models.AnalysisMaster {
	byID := make(map[int32]*models.AnalysisMaster, len(all))
	for i := range all {
		byID[all[i].ID] = &all[i]
	}
	// roots: nodes without a parent, or whose parent is missing from the set
	rootIDs := make([]int32, 0, len(all))
	for i := range all {
		if all[i].ParentID == nil {
			rootIDs = append(rootIDs, all[i].ID)
			continue
		}
		if _, ok := byID[*all[i].ParentID]; !ok {
			rootIDs = append(rootIDs, all[i].ID)
		}
	}
	for i := len(all) - 1; i >= 0; i-- {
		node := &all[i]
		if node.ParentID == nil {
			continue
		}
		parent, ok := byID[*node.ParentID]
		if !ok {
			continue
		}
		// prepend so siblings stay in code-ascending order
		parent.Children = append([]models.AnalysisMaster{*node}, parent.Children...)
	}
	roots := make([]models.AnalysisMaster, 0, len(rootIDs))
	for _, id := range rootIDs {
		roots = append(roots, *byID[id])
	}
	return roots
}

// SearchAHSP searches system AHSP items by name/ahspCode ILIKE (§6.6 search).
func (r *AnalysisMasterRepo) SearchAHSP(ctx context.Context, q, category string, limit int) ([]models.AnalysisMaster, error) {
	args := []any{"%" + q + "%"}
	query := `SELECT id, code, name, level, "parentId", unit, "unitPrice", category, "isGlobal", "userId",
		"isSystem", "ahspCode", "ahspSheet", "generalCost", "createdAt", "updatedAt"
		FROM analysis_masters WHERE "isSystem" = true AND "deletedAt" IS NULL AND (name ILIKE $1 OR "ahspCode" ILIKE $1)`
	if category != "" && category != "custom" {
		args = append(args, category)
		query += ` AND category = $` + strconv.Itoa(len(args))
	}
	if limit <= 0 {
		limit = 20
	}
	args = append(args, limit)
	query += ` ORDER BY similarity(name, $1) DESC, name ASC LIMIT $` + strconv.Itoa(len(args))
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AnalysisMaster
	for rows.Next() {
		m, err := scanAnalysisMaster(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}
