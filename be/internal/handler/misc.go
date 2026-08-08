package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/repository"
)

type DashboardHandler struct {
	pool *pgxpool.Pool
	repo *repository.DashboardRepo
}

func NewDashboardHandler(pool *pgxpool.Pool, repo *repository.DashboardRepo) *DashboardHandler {
	return &DashboardHandler{pool: pool, repo: repo}
}

func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	stats, err := h.repo.Stats(r.Context(), u.UserID, u.Role == models.RoleAdmin)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"stats": stats, "recentProjects": stats.RecentProjects})
}

type FeedbackHandler struct{ repo *repository.FeedbackRepo }

func NewFeedbackHandler(repo *repository.FeedbackRepo) *FeedbackHandler {
	return &FeedbackHandler{repo: repo}
}

func (h *FeedbackHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	out, err := h.repo.ListByUser(r.Context(), u.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"feedback": out})
}

func (h *FeedbackHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if len(in.Message) < 10 {
		writeError(w, http.StatusBadRequest, "Pesan minimal 10 karakter")
		return
	}
	subject := in.Subject
	if subject == "" {
		subject = "other"
	}
	f, err := h.repo.Create(r.Context(), u.UserID, subject, in.Message)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"feedback": f})
}

type AuditLogHandler struct{ repo *repository.AuditLogRepo }

func NewAuditLogHandler(repo *repository.AuditLogRepo) *AuditLogHandler {
	return &AuditLogHandler{repo: repo}
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	f := repository.ListAuditFilter{Action: r.URL.Query().Get("action"), EntityType: r.URL.Query().Get("entityType"), IsAdmin: u.Role == models.RoleAdmin}
	if !f.IsAdmin {
		f.UserID = &u.UserID
	}
	if pid := r.URL.Query().Get("projectId"); pid != "" {
		if v, err := atoi32(pid); err == nil {
			f.ProjectID = &v
		}
	}
	out, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type NewsHandler struct{ repo *repository.NewsRepo }

func NewNewsHandler(repo *repository.NewsRepo) *NewsHandler {
	return &NewsHandler{repo: repo}
}

func (h *NewsHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var out []models.News
	var err error
	if u.Role == models.RoleAdmin {
		out, err = h.repo.ListAll(r.Context())
	} else {
		out, err = h.repo.ListActive(r.Context())
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"news": out})
}

func (h *NewsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Title == "" || in.Content == "" {
		writeError(w, http.StatusBadRequest, "title dan content wajib diisi")
		return
	}
	n, err := h.repo.Create(r.Context(), in.Title, in.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *NewsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	var in struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	n, err := h.repo.Update(r.Context(), id, in.Title, in.Content)
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *NewsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

type MonitoringHandler struct {
	pool *pgxpool.Pool
}

func NewMonitoringHandler(pool *pgxpool.Pool) *MonitoringHandler {
	return &MonitoringHandler{pool: pool}
}

func (h *MonitoringHandler) List(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("projectId")
	if pidStr == "" {
		writeError(w, http.StatusBadRequest, "projectId wajib diisi")
		return
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid projectId")
		return
	}
	if _, _, found, ok := auth.ProjectAccess(r.Context(), h.pool, int32(pid), auth.AccessView); !ok {
		if !found {
			writeError(w, http.StatusNotFound, "Project tidak ditemukan")
		} else {
			writeError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}
	rows, err := h.pool.Query(r.Context(), `
		SELECT w.id, w."description", w.volume::text, w.unit, w.category, w.progress,
			(SELECT l."createdAt" FROM work_item_progress_logs l
			 WHERE l."workItemId" = w.id
			 ORDER BY l."createdAt" DESC, l.id DESC LIMIT 1) AS "lastUpdated"
		FROM work_items w WHERE w."projectId" = $1 AND w."deletedAt" IS NULL
		ORDER BY w.category, w.id`, pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type monItem struct {
		ID          int32   `json:"id"`
		Description string  `json:"description"`
		Volume      string  `json:"volume"`
		Unit        string  `json:"unit"`
		Progress    int     `json:"progress"`
		LastUpdated *string `json:"lastUpdated"`
	}
	grouped := map[string][]monItem{}
	var order []string
	for rows.Next() {
		var id int32
		var description, vol, unit, kat string
		var progress int
		var lastUpdated sql.NullString
		if err := rows.Scan(&id, &description, &vol, &unit, &kat, &progress, &lastUpdated); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if _, exists := grouped[kat]; !exists {
			order = append(order, kat)
		}
		grouped[kat] = append(grouped[kat], monItem{
			ID:          id,
			Description: description,
			Volume:      vol,
			Unit:        unit,
			Progress:    progress,
			LastUpdated: nullStr(lastUpdated),
		})
	}
	type monCategory struct {
		Category string    `json:"category"`
		Items    []monItem `json:"items"`
	}
	var out []monCategory
	for _, k := range order {
		out = append(out, monCategory{Category: k, Items: grouped[k]})
	}
	writeJSON(w, http.StatusOK, map[string]any{"monitoring": out})
}
