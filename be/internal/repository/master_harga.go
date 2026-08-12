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

type PriceMasterRepo struct {
	pool  database.Pool
	cache *cache.Cache
}

func NewPriceMasterRepo(pool database.Pool) *PriceMasterRepo {
	return &PriceMasterRepo{pool: pool, cache: cache.New(60 * time.Second)}
}

type ListPriceMasterFilter struct {
	UserID   int32
	Type     string
	Search   string
	IsGlobal *bool
}

func (f ListPriceMasterFilter) cacheKey() string {
	isGlobal := "nil"
	if f.IsGlobal != nil {
		isGlobal = strconv.FormatBool(*f.IsGlobal)
	}
	return fmt.Sprintf("price|u:%d|type:%s|search:%s|global:%s", f.UserID, f.Type, f.Search, isGlobal)
}

func (r *PriceMasterRepo) List(ctx context.Context, f ListPriceMasterFilter) ([]models.PriceMaster, error) {
	if v, ok := r.cache.Get(f.cacheKey()); ok {
		return v.([]models.PriceMaster), nil
	}
	q := `SELECT id, name, unit, price, type, "isGlobal", "userId", "ahspCode", "isSystem", "createdAt", "updatedAt"
		FROM price_masters WHERE ("userId" = $1 OR "isGlobal" = true OR "isSystem" = true) AND "deletedAt" IS NULL`
	args := []any{f.UserID}
	if f.Type != "" {
		args = append(args, f.Type)
		q += ` AND type = $` + strconv.Itoa(len(args))
	}
	if f.IsGlobal != nil {
		args = append(args, *f.IsGlobal)
		q += ` AND "isGlobal" = $` + strconv.Itoa(len(args))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += ` AND name ILIKE $` + strconv.Itoa(len(args))
	}
	q += ` ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.PriceMaster
	for rows.Next() {
		m, err := scanPriceMaster(rows)
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

func scanPriceMaster(s rowScanner) (*models.PriceMaster, error) {
	var m models.PriceMaster
	var price, ahspCode sql.NullString
	var userID sql.NullInt32
	if err := s.Scan(&m.ID, &m.Name, &m.Unit, &price, &m.Type, &m.IsGlobal, &userID, &ahspCode, &m.IsSystem, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	m.Price = scanDec(price.String)
	m.UserID = i32Ptr(userID)
	m.AHSPCode = strPtr(ahspCode)
	return &m, nil
}

func (r *PriceMasterRepo) Get(ctx context.Context, id int32) (*models.PriceMaster, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, unit, price, type, "isGlobal", "userId", "ahspCode", "isSystem", "createdAt", "updatedAt"
		FROM price_masters WHERE id = $1 AND "deletedAt" IS NULL`, id)
	return scanPriceMaster(row)
}

type CreatePriceMasterInput struct {
	Name     string
	Unit     string
	Price    decimal.Decimal
	Type     models.ComponentType
	IsGlobal bool
	UserID   *int32
	IsSystem bool
}

func (r *PriceMasterRepo) Create(ctx context.Context, in CreatePriceMasterInput) (*models.PriceMaster, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO price_masters (name, unit, price, type, "isGlobal", "userId", "isSystem")
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, name, unit, price, type, "isGlobal", "userId", "ahspCode", "isSystem", "createdAt", "updatedAt"`,
		in.Name, in.Unit, decArg(in.Price), in.Type, in.IsGlobal, in.UserID, in.IsSystem)
	m, err := scanPriceMaster(row)
	if err != nil {
		return nil, err
	}
	r.cache.Clear()
	return m, nil
}

func (r *PriceMasterRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `UPDATE price_masters SET "deletedAt" = NOW() WHERE id = $1`, id)
	if err == nil {
		r.cache.Clear()
	}
	return err
}

type UpdatePriceMasterInput struct {
	Name  *string
	Unit  *string
	Price *decimal.Decimal
	Type  *models.ComponentType
}

func (r *PriceMasterRepo) Update(ctx context.Context, id int32, in UpdatePriceMasterInput) (*models.PriceMaster, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE price_masters SET
			name = COALESCE($2, name),
			unit = COALESCE($3, unit),
			price = COALESCE($4, price),
			type = COALESCE($5, type),
			"updatedAt" = NOW()
		WHERE id = $1
		RETURNING id, name, unit, price, type, "isGlobal", "userId", "ahspCode", "isSystem", "createdAt", "updatedAt"`,
		id, in.Name, in.Unit, decPtrArg(in.Price), in.Type)
	m, err := scanPriceMaster(row)
	if err != nil {
		return nil, err
	}
	r.cache.Clear()
	return m, nil
}

// GetMany loads a set of price_masters rows by ID (used by drift detection).
func (r *PriceMasterRepo) GetMany(ctx context.Context, ids []int32) (map[int32]decimal.Decimal, error) {
	if len(ids) == 0 {
		return map[int32]decimal.Decimal{}, nil
	}
	out := make(map[int32]decimal.Decimal, len(ids))
	rows, err := r.pool.Query(ctx, `SELECT id, price::text FROM price_masters WHERE id = ANY($1) AND "deletedAt" IS NULL`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int32
		var hs string
		if err := rows.Scan(&id, &hs); err != nil {
			return nil, err
		}
		out[id] = scanDec(hs)
	}
	return out, rows.Err()
}
