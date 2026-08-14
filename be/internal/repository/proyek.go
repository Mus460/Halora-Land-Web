package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/shopspring/decimal"

	"github.com/halora-land/halora-be/internal/models"
)

// ProjectRepo handles projects + project_team persistence.
type ProjectRepo struct{ pool database.Pool }

func NewProjectRepo(pool database.Pool) *ProjectRepo { return &ProjectRepo{pool: pool} }

type ListProjectFilter struct {
	UserID  int32
	Search  string
	Type    string
	IsAdmin bool
}

func (r *ProjectRepo) List(ctx context.Context, f ListProjectFilter) ([]models.Project, error) {
	var q string
	var args []any
	if f.IsAdmin {
		q = `SELECT id, "userId", "name", location, type, "isPitching", "isDone", "contractValue", "buildingArea", "timelineMonths", "timelineDays", "createdAt", "updatedAt" FROM projects WHERE "deletedAt" IS NULL`
	} else {
		q = `SELECT DISTINCT p.id, p."userId", p."name", p.location, p.type, p."isPitching", p."isDone", p."contractValue", p."buildingArea", p."timelineMonths", p."timelineDays", p."createdAt", p."updatedAt"
			FROM projects p LEFT JOIN project_team tp ON tp."projectId" = p.id
			WHERE (p."userId" = $1 OR tp."userId" = $1) AND p."deletedAt" IS NULL`
		args = append(args, f.UserID)
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += fmt.Sprintf(` AND "name" ILIKE $%d`, len(args))
	}
	if f.Type != "" {
		args = append(args, f.Type)
		q += fmt.Sprintf(` AND type = $%d`, len(args))
	}
	q += ` ORDER BY "createdAt" DESC`
	return r.scanList(ctx, q, args...)
}

func (r *ProjectRepo) scanList(ctx context.Context, q string, args ...any) ([]models.Project, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Project{}
	for rows.Next() {
		p, err := scanProjectRow(rows)
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

func scanProjectRow(s rowScanner) (*models.Project, error) {
	var p models.Project
	var location, contractValue, buildingArea sql.NullString
	if err := s.Scan(&p.ID, &p.UserID, &p.Name, &location, &p.Type, &p.IsPitching, &p.IsDone, &contractValue, &buildingArea, &p.TimelineMonths, &p.TimelineDays, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Location = strPtr(location)
	p.ContractValue = scanDecPtr(contractValue)
	p.BuildingArea = scanDecPtr(buildingArea)
	return &p, nil
}

func (r *ProjectRepo) Get(ctx context.Context, id int32) (*models.Project, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, "userId", "name", location, type, "isPitching", "isDone", "contractValue", "buildingArea", "timelineMonths", "timelineDays", "createdAt", "updatedAt"
		FROM projects WHERE id = $1 AND "deletedAt" IS NULL`, id)
	p, err := scanProjectRow(row)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type ProjectDetailUser struct {
	ID       int32  `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

type ProjectDetailTeam struct {
	ID   int32             `json:"id"`
	Role models.TeamRole   `json:"role"`
	User ProjectDetailUser `json:"user"`
}

type ProjectDetailWorkItem struct {
	ID          int32               `json:"id"`
	Description string              `json:"description"`
	Volume      decimal.Decimal     `json:"volume"`
	Unit        string              `json:"unit"`
	UnitPrice   decimal.Decimal     `json:"unitPrice"`
	TotalCost   decimal.Decimal     `json:"totalCost"`
	Category    models.WorkCategory `json:"category"`
}

type ProjectDetailCount struct {
	WorkItem int32 `json:"work_items"`
	Recap    int32 `json:"recaps"`
	Invoice  int32 `json:"invoices"`
}

type ProjectDetail struct {
	models.Project
	User        ProjectDetailUser       `json:"user"`
	ProjectTeam []ProjectDetailTeam     `json:"projectTeam"`
	WorkItem    []ProjectDetailWorkItem `json:"work_items"`
	Count       ProjectDetailCount      `json:"_count"`
}

func (r *ProjectRepo) GetDetail(ctx context.Context, id int32) (*ProjectDetail, error) {
	p, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &ProjectDetail{
		Project:     *p,
		ProjectTeam: []ProjectDetailTeam{},
		WorkItem:    []ProjectDetailWorkItem{},
	}

	var owner ProjectDetailUser
	err = r.pool.QueryRow(ctx,
		`SELECT id, "fullName", email FROM users WHERE id = $1`, p.UserID).
		Scan(&owner.ID, &owner.FullName, &owner.Email)
	if err == nil {
		detail.User = owner
	}

	rows, err := r.pool.Query(ctx, `
		SELECT tp.id, tp.role, u.id, u."fullName", u.email
		FROM project_team tp JOIN users u ON u.id = tp."userId"
		WHERE tp."projectId" = $1 ORDER BY tp.id`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t ProjectDetailTeam
			if err := rows.Scan(&t.ID, &t.Role, &t.User.ID, &t.User.FullName, &t.User.Email); err == nil {
				detail.ProjectTeam = append(detail.ProjectTeam, t)
			}
		}
	}

	pRows, err := r.pool.Query(ctx, `
		SELECT id, "description", volume::text, unit, "unitPrice"::text, "totalCost"::text, category
		FROM work_items WHERE "projectId" = $1 AND "deletedAt" IS NULL ORDER BY id DESC LIMIT 10`, id)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var pk ProjectDetailWorkItem
			var vol, hs, tb string
			if err := pRows.Scan(&pk.ID, &pk.Description, &vol, &pk.Unit, &hs, &tb, &pk.Category); err == nil {
				pk.Volume = scanDec(vol)
				pk.UnitPrice = scanDec(hs)
				pk.TotalCost = scanDec(tb)
				detail.WorkItem = append(detail.WorkItem, pk)
			}
		}
	}

	var workItemCount, recapCount, invoiceCount int32
	r.pool.QueryRow(ctx, `SELECT count(*) FROM work_items WHERE "projectId" = $1 AND "deletedAt" IS NULL`, id).Scan(&workItemCount)
	r.pool.QueryRow(ctx, `SELECT count(*) FROM recaps WHERE "projectId" = $1`, id).Scan(&recapCount)
	r.pool.QueryRow(ctx, `SELECT count(*) FROM invoices WHERE "projectId" = $1`, id).Scan(&invoiceCount)
	detail.Count = ProjectDetailCount{WorkItem: workItemCount, Recap: recapCount, Invoice: invoiceCount}

	return detail, nil
}

type CreateProjectInput struct {
	UserID         int32
	Name           string
	Location       *string
	Type           models.ProjectType
	IsPitching     bool
	IsDone         bool
	ContractValue  *decimal.Decimal
	BuildingArea   *decimal.Decimal
	TimelineMonths int
	TimelineDays   int
}

func (r *ProjectRepo) Create(ctx context.Context, in CreateProjectInput) (*models.Project, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO projects ("userId", "name", location, type, "isPitching", "isDone", "contractValue", "buildingArea", "timelineMonths", "timelineDays")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, "userId", "name", location, type, "isPitching", "isDone", "contractValue", "buildingArea", "timelineMonths", "timelineDays", "createdAt", "updatedAt"`,
		in.UserID, in.Name, in.Location, in.Type, in.IsPitching, in.IsDone, decPtrArg(in.ContractValue), decPtrArg(in.BuildingArea), in.TimelineMonths, in.TimelineDays)
	return scanProjectRow(row)
}

// ImportedWorkItem is a BOQ row ready for work_items insertion.
type ImportedWorkItem struct {
	Category    models.WorkCategory
	Description string
	Volume      decimal.Decimal
	Unit        string
	UnitPrice   decimal.Decimal
	TotalCost   decimal.Decimal
}

// ImportedRecap is one row of the BOQ REKAP PER DIVISI section.
type ImportedRecap struct {
	Category    string
	Description string
	Amount      decimal.Decimal
}

// ImportBOQ creates a project together with its work_items and recaps inside a
// single transaction (used when creating a project from a BOQ/RAB file).
func (r *ProjectRepo) ImportBOQ(ctx context.Context, in CreateProjectInput, items []ImportedWorkItem, divisions []ImportedRecap) (*models.Project, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	p, err := scanProjectRow(tx.QueryRow(ctx, `
		INSERT INTO projects ("userId", "name", location, type, "isPitching", "isDone", "contractValue", "buildingArea", "timelineMonths", "timelineDays")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, "userId", "name", location, type, "isPitching", "isDone", "contractValue", "buildingArea", "timelineMonths", "timelineDays", "createdAt", "updatedAt"`,
		in.UserID, in.Name, in.Location, in.Type, in.IsPitching, in.IsDone, decPtrArg(in.ContractValue), decPtrArg(in.BuildingArea), in.TimelineMonths, in.TimelineDays))
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		if it.Description == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO work_items ("projectId", category, "description", volume, unit, "unitPrice", "totalCost", "calculationMethod")
			VALUES ($1,$2,$3,$4,$5,$6,$7,'manual')`,
			p.ID, it.Category, it.Description, decArg(it.Volume), it.Unit, decArg(it.UnitPrice), decArg(it.TotalCost)); err != nil {
			return nil, err
		}
	}

	for i, d := range divisions {
		if d.Category == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO recaps ("projectId", category, description, sequence, margin)
			VALUES ($1,$2,$3,$4,NULL)`,
			p.ID, d.Category, d.Description, i+1); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

type UpdateProjectInput struct {
	Name           *string             `json:"name"`
	Location       *string             `json:"location"`
	Type           *models.ProjectType `json:"type"`
	IsPitching     *bool               `json:"isPitching"`
	IsDone         *bool               `json:"isDone"`
	ContractValue  *decimal.Decimal    `json:"contractValue"`
	BuildingArea   *decimal.Decimal    `json:"buildingArea"`
	TimelineMonths *int                `json:"timelineMonths"`
	TimelineDays   *int                `json:"timelineDays"`
}

func (r *ProjectRepo) Update(ctx context.Context, id int32, in UpdateProjectInput) (*models.Project, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE projects SET
			"name" = COALESCE($2, "name"),
			location = COALESCE($3, location),
			type = COALESCE($4, type),
			"isPitching" = COALESCE($5, "isPitching"),
			"isDone" = COALESCE($6, "isDone"),
			"contractValue" = COALESCE($7, "contractValue"),
			"buildingArea" = COALESCE($8, "buildingArea"),
			"timelineMonths" = COALESCE($9, "timelineMonths"),
			"timelineDays" = COALESCE($10, "timelineDays"),
			"updatedAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, "userId", "name", location, type, "isPitching", "isDone", "contractValue", "buildingArea", "timelineMonths", "timelineDays", "createdAt", "updatedAt"`,
		id, in.Name, in.Location, in.Type, in.IsPitching, in.IsDone, decPtrArg(in.ContractValue), decPtrArg(in.BuildingArea), in.TimelineMonths, in.TimelineDays)
	return scanProjectRow(row)
}

// Delete soft-deletes a projects and its work_items rows in one transaction.
func (r *ProjectRepo) Delete(ctx context.Context, id int32) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE projects SET "deletedAt" = NOW() WHERE id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE work_items SET "deletedAt" = NOW() WHERE "projectId" = $1 AND "deletedAt" IS NULL`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SummaryProject is a lightweight projection used by the recaps/RAB rollup.
type SummaryProject struct {
	ID            int32
	Name          string
	Location      *string
	ContractValue *decimal.Decimal
	BuildingArea  *decimal.Decimal
}

func (r *ProjectRepo) Summary(ctx context.Context, id int32) (*SummaryProject, error) {
	p, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SummaryProject{ID: p.ID, Name: p.Name, Location: p.Location, ContractValue: p.ContractValue, BuildingArea: p.BuildingArea}, nil
}
