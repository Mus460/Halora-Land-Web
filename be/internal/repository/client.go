package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/halora-land/halora-be/internal/cache"
	"github.com/halora-land/halora-be/internal/database"
	"github.com/halora-land/halora-be/internal/models"
)

type ClientRepo struct {
	pool  database.Pool
	cache *cache.Cache
}

func NewClientRepo(pool database.Pool) *ClientRepo {
	return &ClientRepo{pool: pool, cache: cache.New(60 * time.Second)}
}

func (r *ClientRepo) List(ctx context.Context, userID int32, search string) ([]models.Client, error) {
	key := fmt.Sprintf("client|u:%d|s:%s", userID, search)
	if v, ok := r.cache.Get(key); ok {
		return v.([]models.Client), nil
	}
	q := `SELECT id, name, address, contact, "userId", "createdAt", "updatedAt"
		FROM clients WHERE "userId" = $1 AND "deletedAt" IS NULL`
	args := []any{userID}
	if search != "" {
		args = append(args, "%"+search+"%")
		q += ` AND (name ILIKE $2 OR COALESCE(address, '') ILIKE $2)`
	}
	q += ` ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Client
	for rows.Next() {
		m, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	r.cache.Set(key, out)
	return out, nil
}

func scanClient(s rowScanner) (*models.Client, error) {
	var m models.Client
	var address, contact sql.NullString
	if err := s.Scan(&m.ID, &m.Name, &address, &contact, &m.UserID, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	m.Address = strPtr(address)
	m.Contact = strPtr(contact)
	return &m, nil
}

type CreateClientInput struct {
	Name    string
	Address *string
	Contact *string
	UserID  int32
}

func (r *ClientRepo) Create(ctx context.Context, in CreateClientInput) (*models.Client, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO clients (name, address, contact, "userId")
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, address, contact, "userId", "createdAt", "updatedAt"`,
		in.Name, in.Address, in.Contact, in.UserID)
	m, err := scanClient(row)
	if err != nil {
		return nil, err
	}
	r.cache.Clear()
	return m, nil
}

type UpdateClientInput struct {
	Name    *string
	Address *string
	Contact *string
}

func (r *ClientRepo) Update(ctx context.Context, id, userID int32, in UpdateClientInput) (*models.Client, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE clients SET
			name = COALESCE($3, name),
			address = COALESCE($4, address),
			contact = COALESCE($5, contact),
			"updatedAt" = NOW()
		WHERE id = $1 AND "userId" = $2 AND "deletedAt" IS NULL
		RETURNING id, name, address, contact, "userId", "createdAt", "updatedAt"`,
		id, userID, in.Name, in.Address, in.Contact)
	m, err := scanClient(row)
	if err != nil {
		return nil, err
	}
	r.cache.Clear()
	return m, nil
}

func (r *ClientRepo) Delete(ctx context.Context, id, userID int32) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE clients SET "deletedAt" = NOW()
		WHERE id = $1 AND "userId" = $2 AND "deletedAt" IS NULL`, id, userID)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() > 0 {
		r.cache.Clear()
	}
	return tag.RowsAffected() > 0, nil
}