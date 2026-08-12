package handler

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/halora-land/halora-be/internal/database"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/boq"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

type ProjectHandler struct {
	pool database.Pool
	repo *repository.ProjectRepo
}

func NewProjectHandler(pool database.Pool, repo *repository.ProjectRepo) *ProjectHandler {
	return &ProjectHandler{pool: pool, repo: repo}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	f := repository.ListProjectFilter{UserID: u.UserID, Search: r.URL.Query().Get("search"), Type: r.URL.Query().Get("type")}
	if u.Role == models.RoleAdmin {
		f.IsAdmin = true
	}
	out, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": out})
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.repo.GetDetail(r.Context(), id)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	if !h.canView(r, &models.Project{ID: p.ID, UserID: p.UserID}) {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": p})
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		Name           string           `json:"name"`
		Location       *string          `json:"location"`
		Type           string           `json:"type"`
		IsPitching     bool             `json:"isPitching"`
		IsDone         bool             `json:"isDone"`
		ContractValue  *decimal.Decimal `json:"contractValue"`
		TimelineMonths int              `json:"timelineMonths"`
		TimelineDays   int              `json:"timelineDays"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" {
		writeError(w, http.StatusBadRequest, "name wajib diisi")
		return
	}
	if in.TimelineMonths < 0 || in.TimelineDays < 0 {
		writeError(w, http.StatusBadRequest, "timeline tidak boleh negatif")
		return
	}
	ptype := models.ProjectTypeBuilding
	if in.Type == "infra" {
		ptype = models.ProjectTypeInfrastructure
	}
	var nk *decimal.Decimal
	if in.ContractValue != nil && !in.ContractValue.IsZero() {
		nk = in.ContractValue
	}
	p, err := h.repo.Create(r.Context(), repository.CreateProjectInput{
		UserID: u.UserID, Name: in.Name, Location: in.Location, Type: ptype,
		IsPitching: in.IsPitching, IsDone: in.IsDone, ContractValue: nk,
		TimelineMonths: in.TimelineMonths, TimelineDays: in.TimelineDays,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"projects": p})
}

// Import creates a project from an uploaded BOQ/RAB xlsx (multipart/form-data
// with a "boq" file and the standard project form fields). It parses the file,
// then creates the project together with its work_items and recaps.
func (h *ProjectHandler) Import(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "boq: gagal membaca form upload")
		return
	}
	file, _, err := r.FormFile("boq")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file BOQ wajib diunggah (field: boq)")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		writeError(w, http.StatusBadRequest, "gagal membuka file BOQ: "+err.Error())
		return
	}
	defer f.Close()
	doc, err := boq.Parse(f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = doc.Title
	}
	if name == "" {
		writeError(w, http.StatusBadRequest, "name wajib diisi")
		return
	}
	tMonths, _ := strconv.Atoi(r.FormValue("timelineMonths"))
	tDays, _ := strconv.Atoi(r.FormValue("timelineDays"))
	if tMonths < 0 || tDays < 0 {
		writeError(w, http.StatusBadRequest, "timeline tidak boleh negatif")
		return
	}
	ptype := models.ProjectTypeBuilding
	if r.FormValue("type") == "infra" {
		ptype = models.ProjectTypeInfrastructure
	}
	var location *string
	if l := strings.TrimSpace(r.FormValue("location")); l != "" {
		location = &l
	}
	var cv *decimal.Decimal
	if s := strings.TrimSpace(r.FormValue("contractValue")); s != "" {
		if d, err := decimal.NewFromString(s); err == nil && !d.IsZero() {
			cv = &d
		}
	}
	if cv == nil && !doc.Total.IsZero() {
		d := doc.Total
		cv = &d
	}

	items := make([]repository.ImportedWorkItem, 0, len(doc.Items))
	for _, it := range doc.Items {
		items = append(items, repository.ImportedWorkItem{
			Category: it.Category, Description: it.Description, Volume: it.Volume,
			Unit: it.Unit, UnitPrice: it.UnitPrice, TotalCost: it.TotalCost,
		})
	}
	divisions := make([]repository.ImportedRecap, 0, len(doc.Divisions))
	for _, d := range doc.Divisions {
		divisions = append(divisions, repository.ImportedRecap{
			Category: d.Letter, Description: d.Description, Amount: d.Amount,
		})
	}

	p, err := h.repo.ImportBOQ(r.Context(), repository.CreateProjectInput{
		UserID: u.UserID, Name: name, Location: location, Type: ptype,
		ContractValue: cv, TimelineMonths: tMonths, TimelineDays: tDays,
	}, items, divisions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"projects": p,
		"imported": map[string]any{"workItems": len(items), "recaps": len(divisions)},
	})
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, id, auth.AccessEdit); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	var in repository.UpdateProjectInput
	if !decodeJSON(w, r, &in) {
		return
	}
	updated, err := h.repo.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": updated})
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, id, auth.AccessOwner); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *ProjectHandler) canView(r *http.Request, p *models.Project) bool {
	_, _, _, ok := auth.ProjectAccess(r.Context(), h.pool, p.ID, auth.AccessView)
	return ok
}
