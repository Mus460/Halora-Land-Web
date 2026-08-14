package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/database"
	"github.com/halora-land/halora-be/internal/env"
)

func seed(cfg *env.Config) error {
	ctx := context.Background()
	pool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}
	if err := seedDB(ctx, pool); err != nil {
		return err
	}
	log.Println("seed complete")
	return nil
}

// seedDB inserts the baseline seed users + a sample project + price_masters
// (port of prisma/seed.ts, ARCHITECTURE.md §5.3). All auth is local DB auth
// (bcrypt password hashes) — no external provider required.
func seedDB(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	users := []struct {
		name, email, role, pwd string
		isDemo                 bool
	}{
		{"Admin Halora", "admin@haloraland.id", "ADMIN", "admin123", false},
		{"Budi Santoso", "budi@example.com", "USER", "password123", false},
		{"Demo User", "demo@haloraland.id", "DEMO", "demo123", true},
	}
	var budiID int32
	for i, u := range users {
		hash, err := auth.HashPassword(u.pwd)
		if err != nil {
			return fmt.Errorf("seed hash pwd: %w", err)
		}
		var id int32
		err = tx.QueryRow(ctx, `
			INSERT INTO users ("fullName", email, role, "accountType", "isDemo", "passwordHash")
			VALUES ($1,$2,$3,'free',$4,$5)
			ON CONFLICT (email) DO UPDATE SET "fullName" = EXCLUDED."fullName", "passwordHash" = EXCLUDED."passwordHash"
			RETURNING id`, u.name, u.email, u.role, u.isDemo, string(hash)).Scan(&id)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.email, err)
		}
		if i == 1 {
			budiID = id
		}
	}

	var projectID int32
	if err := tx.QueryRow(ctx, `
		INSERT INTO projects ("userId", "name", location, type, "contractValue", "timelineMonths")
		VALUES ($1,'Pembangunan Rumah 2 Lantai','Jakarta Selatan','building',850000000,6)
		RETURNING id`, budiID).Scan(&projectID); err != nil {
		return err
	}

	prices := []struct {
		name, unit, category string
		price                string
	}{
		{"Semen Portland", "zak", "material", "75000"},
		{"Pasir Pasang", "m3", "material", "350000"},
		{"Tukang Batu", "hari", "labor", "150000"},
	}
	for _, p := range prices {
		if _, err := tx.Exec(ctx, `
			INSERT INTO price_masters (name, unit, price, type, "isGlobal", "isSystem")
			VALUES ($1,$2,$3,$4,true,true)`, p.name, p.unit, p.price, p.category); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO work_items ("projectId", category, "description", volume, unit, "unitPrice", "totalCost", "calculationMethod")
		VALUES ($1,'foundation','Galian Tanah Pondasi',25.5,'m3',125000,3187500,'manual')`, projectID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO work_items ("projectId", category, "description", volume, unit, "unitPrice", "totalCost", "calculationMethod")
		VALUES ($1,'wall','Pasang Bata Merah',120,'m2',185000,22200000,'ahsp')`, projectID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
