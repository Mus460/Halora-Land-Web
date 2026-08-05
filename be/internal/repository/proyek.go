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
		q = `SELECT id, "userId", "namaProyek", lokasi, tipe, "isPitching", "nilaiKontrak", timeline, "createdAt", "updatedAt" FROM proyek WHERE 1=1`
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
	out := []models.Proyek{}
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
	if err := s.Scan(&p.ID, &p.UserID, &p.NamaProyek, &lokasi, &p.Tipe, &p.IsPitching, &nilaiKontrak, &timeline, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Lokasi = strPtr(lokasi)
	p.Timeline = strPtr(timeline)
	p.NilaiKontrak = scanDecPtr(nilaiKontrak)
	return &p, nil
}

func (r *ProyekRepo) Get(ctx context.Context, id int32) (*models.Proyek, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, "userId", "namaProyek", lokasi, tipe, "isPitching", "nilaiKontrak", timeline, "createdAt", "updatedAt"
		FROM proyek WHERE id = $1`, id)
	p, err := scanProyekRow(row)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type ProyekDetailUser struct {
	ID           int32  `json:"id"`
	NamaLengkap  string `json:"namaLengkap"`
	Email        string `json:"email"`
}

type ProyekDetailTim struct {
	ID   int32                 `json:"id"`
	Role models.RoleTimProyek  `json:"role"`
	User ProyekDetailUser      `json:"user"`
}

type ProyekDetailPekerjaan struct {
	ID              int32                  `json:"id"`
	UraianPekerjaan string                 `json:"uraianPekerjaan"`
	Volume          decimal.Decimal        `json:"volume"`
	Satuan          string                 `json:"satuan"`
	HargaSatuan     decimal.Decimal        `json:"hargaSatuan"`
	TotalBiaya      decimal.Decimal        `json:"totalBiaya"`
	Kategori        models.KategoriPekerjaan `json:"kategori"`
}

type ProyekDetailCount struct {
	Pekerjaan int32 `json:"pekerjaan"`
	Rekap     int32 `json:"rekap"`
	Invoice   int32 `json:"invoice"`
}

type ProyekDetail struct {
	models.Proyek
	User       ProyekDetailUser        `json:"user"`
	TimProyek  []ProyekDetailTim       `json:"timProyek"`
	Pekerjaan  []ProyekDetailPekerjaan `json:"pekerjaan"`
	Count      ProyekDetailCount       `json:"_count"`
}

func (r *ProyekRepo) GetDetail(ctx context.Context, id int32) (*ProyekDetail, error) {
	p, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &ProyekDetail{
		Proyek:     *p,
		TimProyek:  []ProyekDetailTim{},
		Pekerjaan:  []ProyekDetailPekerjaan{},
	}

	var owner ProyekDetailUser
	err = r.pool.QueryRow(ctx,
		`SELECT id, "namaLengkap", email FROM users WHERE id = $1`, p.UserID).
		Scan(&owner.ID, &owner.NamaLengkap, &owner.Email)
	if err == nil {
		detail.User = owner
	}

	rows, err := r.pool.Query(ctx, `
		SELECT tp.id, tp.role, u.id, u."namaLengkap", u.email
		FROM tim_proyek tp JOIN users u ON u.id = tp."userId"
		WHERE tp."proyekId" = $1 ORDER BY tp.id`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t ProyekDetailTim
			if err := rows.Scan(&t.ID, &t.Role, &t.User.ID, &t.User.NamaLengkap, &t.User.Email); err == nil {
				detail.TimProyek = append(detail.TimProyek, t)
			}
		}
	}

	pRows, err := r.pool.Query(ctx, `
		SELECT id, "uraianPekerjaan", volume::text, satuan, "hargaSatuan"::text, "totalBiaya"::text, kategori
		FROM pekerjaan WHERE "proyekId" = $1 ORDER BY id DESC LIMIT 10`, id)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var pk ProyekDetailPekerjaan
			var vol, hs, tb string
			if err := pRows.Scan(&pk.ID, &pk.UraianPekerjaan, &vol, &pk.Satuan, &hs, &tb, &pk.Kategori); err == nil {
				pk.Volume = scanDec(vol)
				pk.HargaSatuan = scanDec(hs)
				pk.TotalBiaya = scanDec(tb)
				detail.Pekerjaan = append(detail.Pekerjaan, pk)
			}
		}
	}

	var pekerjaanCount, rekapCount, invoiceCount int32
	r.pool.QueryRow(ctx, `SELECT count(*) FROM pekerjaan WHERE "proyekId" = $1`, id).Scan(&pekerjaanCount)
	r.pool.QueryRow(ctx, `SELECT count(*) FROM rekap WHERE "proyekId" = $1`, id).Scan(&rekapCount)
	r.pool.QueryRow(ctx, `SELECT count(*) FROM invoice WHERE "proyekId" = $1`, id).Scan(&invoiceCount)
	detail.Count = ProyekDetailCount{Pekerjaan: pekerjaanCount, Rekap: rekapCount, Invoice: invoiceCount}

	return detail, nil
}

type CreateProyekInput struct {
	UserID       int32
	NamaProyek   string
	Lokasi       *string
	Tipe         models.TipeProyek
	IsPitching   bool
	NilaiKontrak *decimal.Decimal
	Timeline     *string
}

func (r *ProyekRepo) Create(ctx context.Context, in CreateProyekInput) (*models.Proyek, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO proyek ("userId", "namaProyek", lokasi, tipe, "isPitching", "nilaiKontrak", timeline)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, "userId", "namaProyek", lokasi, tipe, "isPitching", "nilaiKontrak", timeline, "createdAt", "updatedAt"`,
		in.UserID, in.NamaProyek, in.Lokasi, in.Tipe, in.IsPitching, decPtrArg(in.NilaiKontrak), in.Timeline)
	return scanProyekRow(row)
}

type UpdateProyekInput struct {
	NamaProyek   *string             `json:"namaProyek"`
	Lokasi       *string             `json:"lokasi"`
	Tipe         *models.TipeProyek  `json:"tipe"`
	IsPitching   *bool               `json:"isPitching"`
	NilaiKontrak *decimal.Decimal    `json:"nilaiKontrak"`
	Timeline     *string             `json:"timeline"`
}

func (r *ProyekRepo) Update(ctx context.Context, id int32, in UpdateProyekInput) (*models.Proyek, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE proyek SET
			"namaProyek" = COALESCE($2, "namaProyek"),
			lokasi = COALESCE($3, lokasi),
			tipe = COALESCE($4, tipe),
			"isPitching" = COALESCE($5, "isPitching"),
			"nilaiKontrak" = COALESCE($6, "nilaiKontrak"),
			timeline = COALESCE($7, timeline),
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, "userId", "namaProyek", lokasi, tipe, "isPitching", "nilaiKontrak", timeline, "createdAt", "updatedAt"`,
		id, in.NamaProyek, in.Lokasi, in.Tipe, in.IsPitching, decPtrArg(in.NilaiKontrak), in.Timeline)
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
