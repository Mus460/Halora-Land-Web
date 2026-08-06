package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/cache"
	"github.com/halora-land/halora-be/internal/models"
)

type MasterHargaRepo struct {
	pool  *pgxpool.Pool
	cache *cache.Cache
}

func NewMasterHargaRepo(pool *pgxpool.Pool) *MasterHargaRepo {
	return &MasterHargaRepo{pool: pool, cache: cache.New(60 * time.Second)}
}

type ListMasterHargaFilter struct {
	UserID   int32
	Kategori string
	Search   string
	IsGlobal *bool
}

func (f ListMasterHargaFilter) cacheKey() string {
	isGlobal := "nil"
	if f.IsGlobal != nil {
		isGlobal = strconv.FormatBool(*f.IsGlobal)
	}
	return fmt.Sprintf("harga|u:%d|kategori:%s|search:%s|global:%s", f.UserID, f.Kategori, f.Search, isGlobal)
}

func (r *MasterHargaRepo) List(ctx context.Context, f ListMasterHargaFilter) ([]models.MasterHarga, error) {
	if v, ok := r.cache.Get(f.cacheKey()); ok {
		return v.([]models.MasterHarga), nil
	}
	q := `SELECT id, nama, satuan, harga, kategori, "isGlobal", "userId", "kodeAHSP", "isSystem", "createdAt", "updatedAt"
		FROM master_harga WHERE ("userId" = $1 OR "isGlobal" = true OR "isSystem" = true) AND "deletedAt" IS NULL`
	args := []any{f.UserID}
	if f.Kategori != "" {
		args = append(args, f.Kategori)
		q += ` AND kategori = $` + strconv.Itoa(len(args))
	}
	if f.IsGlobal != nil {
		args = append(args, *f.IsGlobal)
		q += ` AND "isGlobal" = $` + strconv.Itoa(len(args))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += ` AND nama ILIKE $` + strconv.Itoa(len(args))
	}
	q += ` ORDER BY nama ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.MasterHarga
	for rows.Next() {
		m, err := scanMasterHarga(rows)
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

func scanMasterHarga(s rowScanner) (*models.MasterHarga, error) {
	var m models.MasterHarga
	var harga, kodeAHSP sql.NullString
	var userID sql.NullInt32
	if err := s.Scan(&m.ID, &m.Nama, &m.Satuan, &harga, &m.Kategori, &m.IsGlobal, &userID, &kodeAHSP, &m.IsSystem, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	m.Harga = scanDec(harga.String)
	m.UserID = i32Ptr(userID)
	m.KodeAHSP = strPtr(kodeAHSP)
	return &m, nil
}

func (r *MasterHargaRepo) Get(ctx context.Context, id int32) (*models.MasterHarga, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, nama, satuan, harga, kategori, "isGlobal", "userId", "kodeAHSP", "isSystem", "createdAt", "updatedAt"
		FROM master_harga WHERE id = $1 AND "deletedAt" IS NULL`, id)
	return scanMasterHarga(row)
}

type CreateMasterHargaInput struct {
	Nama     string
	Satuan   string
	Harga    decimal.Decimal
	Kategori models.TipeKomponen
	IsGlobal bool
	UserID   *int32
	IsSystem bool
}

func (r *MasterHargaRepo) Create(ctx context.Context, in CreateMasterHargaInput) (*models.MasterHarga, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO master_harga (nama, satuan, harga, kategori, "isGlobal", "userId", "isSystem")
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, nama, satuan, harga, kategori, "isGlobal", "userId", "kodeAHSP", "isSystem", "createdAt", "updatedAt"`,
		in.Nama, in.Satuan, decArg(in.Harga), in.Kategori, in.IsGlobal, in.UserID, in.IsSystem)
	m, err := scanMasterHarga(row)
	if err != nil {
		return nil, err
	}
	r.cache.Clear()
	return m, nil
}

func (r *MasterHargaRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `UPDATE master_harga SET "deletedAt" = NOW() WHERE id = $1`, id)
	if err == nil {
		r.cache.Clear()
	}
	return err
}

type UpdateMasterHargaInput struct {
	Nama     *string
	Satuan   *string
	Harga    *decimal.Decimal
	Kategori *models.TipeKomponen
}

func (r *MasterHargaRepo) Update(ctx context.Context, id int32, in UpdateMasterHargaInput) (*models.MasterHarga, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE master_harga SET
			nama = COALESCE($2, nama),
			satuan = COALESCE($3, satuan),
			harga = COALESCE($4, harga),
			kategori = COALESCE($5, kategori),
			"updatedAt" = NOW()
		WHERE id = $1
		RETURNING id, nama, satuan, harga, kategori, "isGlobal", "userId", "kodeAHSP", "isSystem", "createdAt", "updatedAt"`,
		id, in.Nama, in.Satuan, decPtrArg(in.Harga), in.Kategori)
	m, err := scanMasterHarga(row)
	if err != nil {
		return nil, err
	}
	r.cache.Clear()
	return m, nil
}

// GetMany loads a set of master_harga rows by ID (used by drift detection).
func (r *MasterHargaRepo) GetMany(ctx context.Context, ids []int32) (map[int32]decimal.Decimal, error) {
	if len(ids) == 0 {
		return map[int32]decimal.Decimal{}, nil
	}
	out := make(map[int32]decimal.Decimal, len(ids))
	rows, err := r.pool.Query(ctx, `SELECT id, harga::text FROM master_harga WHERE id = ANY($1) AND "deletedAt" IS NULL`, ids)
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
