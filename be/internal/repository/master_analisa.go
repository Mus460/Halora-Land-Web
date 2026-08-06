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

type MasterAnalisaRepo struct {
	pool  *pgxpool.Pool
	cache *cache.Cache
}

func NewMasterAnalisaRepo(pool *pgxpool.Pool) *MasterAnalisaRepo {
	return &MasterAnalisaRepo{pool: pool, cache: cache.New(60 * time.Second)}
}

type ListMasterAnalisaFilter struct {
	Level    *int
	ParentID *int32
	Search   string
	IsGlobal *bool
	UserID   int32
}

func (f ListMasterAnalisaFilter) cacheKey() string {
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

func (r *MasterAnalisaRepo) List(ctx context.Context, f ListMasterAnalisaFilter) ([]models.MasterAnalisa, error) {
	if v, ok := r.cache.Get(f.cacheKey()); ok {
		return v.([]models.MasterAnalisa), nil
	}
	q := `SELECT id, kode, nama, level, "parentId", satuan, "hargaSatuan", kategori, "isGlobal", "userId",
		"isSystem", "ahspKode", "ahspSheet", "biayaUmum", "createdAt", "updatedAt"
		FROM master_analisa WHERE ("userId" = $1 OR "isGlobal" = true OR "isSystem" = true) AND "deletedAt" IS NULL`
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
		q += ` AND (nama ILIKE $` + strconv.Itoa(len(args))
		args = append(args, "%"+f.Search+"%")
		q += ` OR "ahspKode" ILIKE $` + strconv.Itoa(len(args)) + `)`
	}
	q += ` ORDER BY level ASC, kode ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.MasterAnalisa
	for rows.Next() {
		m, err := scanMasterAnalisa(rows)
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

func scanMasterAnalisa(s rowScanner) (*models.MasterAnalisa, error) {
	var m models.MasterAnalisa
	var parentID sql.NullInt32
	var satuan, kategori, ahspKode, ahspSheet, hs, biayaUmum sql.NullString
	var userID sql.NullInt32
	if err := s.Scan(&m.ID, &m.Kode, &m.Nama, &m.Level, &parentID, &satuan, &hs, &kategori,
		&m.IsGlobal, &userID, &m.IsSystem, &ahspKode, &ahspSheet, &biayaUmum, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	m.ParentID = i32Ptr(parentID)
	m.Satuan = strPtr(satuan)
	m.HargaSatuan = scanDecPtr(hs)
	m.Kategori = strPtr(kategori)
	m.UserID = i32Ptr(userID)
	m.AHSPKode = strPtr(ahspKode)
	m.AHSPSheet = strPtr(ahspSheet)
	m.BiayaUmum = scanDec(biayaUmum.String)
	return &m, nil
}

func (r *MasterAnalisaRepo) Get(ctx context.Context, id int32) (*models.MasterAnalisa, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, kode, nama, level, "parentId", satuan, "hargaSatuan", kategori, "isGlobal", "userId",
		"isSystem", "ahspKode", "ahspSheet", "biayaUmum", "createdAt", "updatedAt"
		FROM master_analisa WHERE id = $1 AND "deletedAt" IS NULL`, id)
	return scanMasterAnalisa(row)
}

type CreateMasterAnalisaInput struct {
	Kode     string
	Nama     string
	Level    int32
	ParentID *int32
	Satuan   *string
	IsGlobal bool
	UserID   *int32
	IsSystem bool
}

func (r *MasterAnalisaRepo) Create(ctx context.Context, in CreateMasterAnalisaInput) (*models.MasterAnalisa, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO master_analisa (kode, nama, level, "parentId", satuan, "isGlobal", "userId", "isSystem")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, kode, nama, level, "parentId", satuan, "hargaSatuan", kategori, "isGlobal", "userId",
		"isSystem", "ahspKode", "ahspSheet", "biayaUmum", "createdAt", "updatedAt"`,
		in.Kode, in.Nama, in.Level, in.ParentID, in.Satuan, in.IsGlobal, in.UserID, in.IsSystem)
	m, err := scanMasterAnalisa(row)
	if err != nil {
		return nil, err
	}
	r.cache.Clear()
	return m, nil
}

func (r *MasterAnalisaRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `UPDATE master_analisa SET "deletedAt" = NOW() WHERE id = $1`, id)
	if err == nil {
		r.cache.Clear()
	}
	return err
}

func (r *MasterAnalisaRepo) HasChildren(ctx context.Context, id int32) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM master_analisa WHERE "parentId" = $1 AND "deletedAt" IS NULL`, id).Scan(&n)
	return n > 0, err
}

// --- rincian_analisa (template breakdown, §3.2) ---

func (r *MasterAnalisaRepo) ListRincian(ctx context.Context, masterAnalisaID int32) ([]models.RincianAnalisa, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "masterAnalisaId", "komponenId", koef, tipe, nama, satuan, "hargaSatuan", "jumlahHarga",
		"kodeReferensi", waktu, urutan, "createdAt", "updatedAt"
		FROM rincian_analisa WHERE "masterAnalisaId" = $1 AND "deletedAt" IS NULL ORDER BY urutan ASC, id ASC`, masterAnalisaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.RincianAnalisa
	for rows.Next() {
		var rin models.RincianAnalisa
		var komponenID sql.NullInt32
		var koef string
		var nama, satuan, hs, jh, kodeRef, waktu sql.NullString
		if err := rows.Scan(&rin.ID, &rin.MasterAnalisaID, &komponenID, &koef, &rin.Tipe, &nama, &satuan, &hs, &jh, &kodeRef, &waktu, &rin.Urutan, &rin.CreatedAt, &rin.UpdatedAt); err != nil {
			return nil, err
		}
		rin.KomponenID = i32Ptr(komponenID)
		rin.Koef = scanDec(koef)
		rin.Nama = strPtr(nama)
		rin.Satuan = strPtr(satuan)
		rin.HargaSatuan = scanDecPtr(hs)
		rin.JumlahHarga = scanDecPtr(jh)
		rin.KodeReferensi = strPtr(kodeRef)
		rin.Waktu = scanDecPtr(waktu)
		out = append(out, rin)
	}
	return out, rows.Err()
}

type CreateRincianInput struct {
	MasterAnalisaID int32                  `json:"masterAnalisaId"`
	KomponenID      *int32                 `json:"komponenId"`
	Koef            decimal.Decimal        `json:"koef"`
	Tipe            models.TipeKomponen    `json:"tipe"`
}

func (r *MasterAnalisaRepo) CreateRincian(ctx context.Context, in CreateRincianInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rincian_analisa ("masterAnalisaId", "komponenId", koef, tipe)
		VALUES ($1,$2,$3,$4)`, in.MasterAnalisaID, in.KomponenID, decArg(in.Koef), in.Tipe)
	return err
}

func (r *MasterAnalisaRepo) DeleteRincian(ctx context.Context, masterAnalisaID, rincianID int32) error {
	_, err := r.pool.Exec(ctx, `UPDATE rincian_analisa SET "deletedAt" = NOW() WHERE id = $1 AND "masterAnalisaId" = $2`, rincianID, masterAnalisaID)
	return err
}

// ListTree fetches all accessible nodes and builds a tree from the root level.
func (r *MasterAnalisaRepo) ListTree(ctx context.Context, f ListMasterAnalisaFilter) ([]models.MasterAnalisa, error) {
	treeFilter := f
	treeFilter.Level = nil
	treeFilter.ParentID = nil
	all, err := r.List(ctx, treeFilter)
	if err != nil {
		return nil, err
	}
	return buildMasterAnalisaTree(all), nil
}

func buildMasterAnalisaTree(all []models.MasterAnalisa) []models.MasterAnalisa {
	byID := make(map[int32]*models.MasterAnalisa, len(all))
	roots := make([]models.MasterAnalisa, 0)
	for i := range all {
		byID[all[i].ID] = &all[i]
	}
	for i := range all {
		node := &all[i]
		if node.ParentID != nil {
			if parent, ok := byID[*node.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
				continue
			}
		}
		roots = append(roots, *node)
	}
	return roots
}

// SearchAHSP searches system AHSP items by nama/ahspKode ILIKE (§6.6 search).
func (r *MasterAnalisaRepo) SearchAHSP(ctx context.Context, q, kategori string, limit int) ([]models.MasterAnalisa, error) {
	args := []any{"%" + q + "%"}
	query := `SELECT id, kode, nama, level, "parentId", satuan, "hargaSatuan", kategori, "isGlobal", "userId",
		"isSystem", "ahspKode", "ahspSheet", "biayaUmum", "createdAt", "updatedAt"
		FROM master_analisa WHERE "isSystem" = true AND "deletedAt" IS NULL AND (nama ILIKE $1 OR "ahspKode" ILIKE $1)`
	if kategori != "" && kategori != "custom" {
		args = append(args, kategori)
		query += ` AND kategori = $` + strconv.Itoa(len(args))
	}
	if limit <= 0 {
		limit = 20
	}
	args = append(args, limit)
	query += ` ORDER BY similarity(nama, $1) DESC, nama ASC LIMIT $` + strconv.Itoa(len(args))
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.MasterAnalisa
	for rows.Next() {
		m, err := scanMasterAnalisa(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}
