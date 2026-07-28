package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

// ProyekRepo handles proyek + tim_proyek persistence.
type ProyekRepo struct{ pool *pgxpool.Pool }

func NewProyekRepo(pool *pgxpool.Pool) *ProyekRepo { return &ProyekRepo{pool: pool} }

type ListProyekFilter struct {
	UserID  int32
	Search  string
	Tipe    string
	IsAdmin bool
}

func (r *ProyekRepo) List(ctx context.Context, f ListProyekFilter) ([]models.Proyek, error) {
	var q string
	var args []any
	if f.IsAdmin {
		q = `SELECT id, "userId", "namaProyek", lokasi, tipe, "nilaiKontrak", timeline, "createdAt", "updatedAt" FROM proyek WHERE 1=1`
	} else {
		q = `SELECT DISTINCT p.id, p."userId", p."namaProyek", p.lokasi, p.tipe, p."nilaiKontrak", p.timeline, p."createdAt", p."updatedAt"
			FROM proyek p LEFT JOIN tim_proyek tp ON tp."proyekId" = p.id
			WHERE (p."userId" = $1 OR tp."userId" = $1)`
		args = append(args, f.UserID)
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += fmt.Sprintf(` AND "namaProyek" ILIKE $%d`, len(args))
	}
	if f.Tipe != "" {
		args = append(args, f.Tipe)
		q += fmt.Sprintf(` AND tipe = $%d`, len(args))
	}
	q += ` ORDER BY "createdAt" DESC`
	return r.scanList(ctx, q, args...)
}

func (r *ProyekRepo) scanList(ctx context.Context, q string, args ...any) ([]models.Proyek, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Proyek
	for rows.Next() {
		p, err := scanProyekRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProyekRow(s rowScanner) (*models.Proyek, error) {
	var p models.Proyek
	var lokasi, timeline, nilaiKontrak sql.NullString
	if err := s.Scan(&p.ID, &p.UserID, &p.NamaProyek, &lokasi, &p.Tipe, &nilaiKontrak, &timeline, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Lokasi = strPtr(lokasi)
	p.Timeline = strPtr(timeline)
	p.NilaiKontrak = scanDecPtr(nilaiKontrak)
	return &p, nil
}

func (r *ProyekRepo) Get(ctx context.Context, id int32) (*models.Proyek, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, "userId", "namaProyek", lokasi, tipe, "nilaiKontrak", timeline, "createdAt", "updatedAt"
		FROM proyek WHERE id = $1`, id)
	p, err := scanProyekRow(row)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type CreateProyekInput struct {
	UserID       int32
	NamaProyek   string
	Lokasi       *string
	Tipe         models.TipeProyek
	NilaiKontrak *decimal.Decimal
	Timeline     *string
}

func (r *ProyekRepo) Create(ctx context.Context, in CreateProyekInput) (*models.Proyek, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO proyek ("userId", "namaProyek", lokasi, tipe, "nilaiKontrak", timeline)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, "userId", "namaProyek", lokasi, tipe, "nilaiKontrak", timeline, "createdAt", "updatedAt"`,
		in.UserID, in.NamaProyek, in.Lokasi, in.Tipe, decPtrArg(in.NilaiKontrak), in.Timeline)
	return scanProyekRow(row)
}

type UpdateProyekInput struct {
	NamaProyek   *string
	Lokasi       *string
	Tipe         *models.TipeProyek
	NilaiKontrak *decimal.Decimal
	Timeline     *string
}

func (r *ProyekRepo) Update(ctx context.Context, id int32, in UpdateProyekInput) (*models.Proyek, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE proyek SET
			"namaProyek" = COALESCE($2, "namaProyek"),
			lokasi = COALESCE($3, lokasi),
			tipe = COALESCE($4, tipe),
			"nilaiKontrak" = COALESCE($5, "nilaiKontrak"),
			timeline = COALESCE($6, timeline),
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, "userId", "namaProyek", lokasi, tipe, "nilaiKontrak", timeline, "createdAt", "updatedAt"`,
		id, in.NamaProyek, in.Lokasi, in.Tipe, decPtrArg(in.NilaiKontrak), in.Timeline)
	return scanProyekRow(row)
}

func (r *ProyekRepo) Delete(ctx context.Context, id int32) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM proyek WHERE id = $1`, id)
	return err
}

// SummaryProyek is a lightweight projection used by the rekap/RAB rollup.
type SummaryProyek struct {
	ID           int32
	NamaProyek   string
	Lokasi       *string
	NilaiKontrak *decimal.Decimal
}

func (r *ProyekRepo) Summary(ctx context.Context, id int32) (*SummaryProyek, error) {
	p, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SummaryProyek{ID: p.ID, NamaProyek: p.NamaProyek, Lokasi: p.Lokasi, NilaiKontrak: p.NilaiKontrak}, nil
}
