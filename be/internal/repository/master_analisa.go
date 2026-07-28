package repository

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

type MasterAnalisaRepo struct{ pool *pgxpool.Pool }

func NewMasterAnalisaRepo(pool *pgxpool.Pool) *MasterAnalisaRepo {
	return &MasterAnalisaRepo{pool: pool}
}

type ListMasterAnalisaFilter struct {
	Level    *int
	ParentID *int32
	Search   string
	IsGlobal *bool
	UserID   int32
}

func (r *MasterAnalisaRepo) List(ctx context.Context, f ListMasterAnalisaFilter) ([]models.MasterAnalisa, error) {
	q := `SELECT id, kode, nama, level, "parentId", satuan, "hargaSatuan", kategori, "isGlobal", "userId",
		"isSystem", "ahspKode", "ahspSheet", "biayaUmum", "createdAt", "updatedAt"
		FROM master_analisa WHERE ("userId" = $1 OR "isGlobal" = true OR "isSystem" = true)`
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
	return out, rows.Err()
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
		FROM master_analisa WHERE id = $1`, id)
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
	return scanMasterAnalisa(row)
}

func (r *MasterAnalisaRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM master_analisa WHERE id = $1`, id)
	return err
}

func (r *MasterAnalisaRepo) HasChildren(ctx context.Context, id int32) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM master_analisa WHERE "parentId" = $1`, id).Scan(&n)
	return n > 0, err
}

// --- rincian_analisa (template breakdown, §3.2) ---

func (r *MasterAnalisaRepo) ListRincian(ctx context.Context, masterAnalisaID int32) ([]models.RincianAnalisa, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, "masterAnalisaId", "komponenId", koef, tipe, nama, satuan, "hargaSatuan", "jumlahHarga",
		"kodeReferensi", urutan, "createdAt", "updatedAt"
		FROM rincian_analisa WHERE "masterAnalisaId" = $1 ORDER BY urutan ASC, id ASC`, masterAnalisaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.RincianAnalisa
	for rows.Next() {
		var rin models.RincianAnalisa
		var komponenID sql.NullInt32
		var koef string
		var nama, satuan, hs, jh, kodeRef sql.NullString
		if err := rows.Scan(&rin.ID, &rin.MasterAnalisaID, &komponenID, &koef, &rin.Tipe, &nama, &satuan, &hs, &jh, &kodeRef, &rin.Urutan, &rin.CreatedAt, &rin.UpdatedAt); err != nil {
			return nil, err
		}
		rin.KomponenID = i32Ptr(komponenID)
		rin.Koef = scanDec(koef)
		rin.Nama = strPtr(nama)
		rin.Satuan = strPtr(satuan)
		rin.HargaSatuan = scanDecPtr(hs)
		rin.JumlahHarga = scanDecPtr(jh)
		rin.KodeReferensi = strPtr(kodeRef)
		out = append(out, rin)
	}
	return out, rows.Err()
}

type CreateRincianInput struct {
	MasterAnalisaID int32
	KomponenID      *int32
	Koef            decimal.Decimal
	Tipe            models.TipeKomponen
}

func (r *MasterAnalisaRepo) CreateRincian(ctx context.Context, in CreateRincianInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rincian_analisa ("masterAnalisaId", "komponenId", koef, tipe)
		VALUES ($1,$2,$3,$4)`, in.MasterAnalisaID, in.KomponenID, decArg(in.Koef), in.Tipe)
	return err
}

func (r *MasterAnalisaRepo) DeleteRincian(ctx context.Context, masterAnalisaID, rincianID int32) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM rincian_analisa WHERE id = $1 AND "masterAnalisaId" = $2`, rincianID, masterAnalisaID)
	return err
}

// SearchAHSP searches system AHSP items by nama/ahspKode ILIKE (§6.6 search).
func (r *MasterAnalisaRepo) SearchAHSP(ctx context.Context, q, kategori string, limit int) ([]models.MasterAnalisa, error) {
	args := []any{"%" + q + "%"}
	query := `SELECT id, kode, nama, level, "parentId", satuan, "hargaSatuan", kategori, "isGlobal", "userId",
		"isSystem", "ahspKode", "ahspSheet", "biayaUmum", "createdAt", "updatedAt"
		FROM master_analisa WHERE "isSystem" = true AND (nama ILIKE $1 OR "ahspKode" ILIKE $1)`
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
