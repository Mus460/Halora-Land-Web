package handler

import (
	"net/http"

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
	writeJSON(w, http.StatusOK, stats)
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
	writeJSON(w, http.StatusOK, out)
}

func (h *FeedbackHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		Message   string `json:"message"`
		Category  string `json:"category"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if len(in.Message) < 10 {
		writeError(w, http.StatusBadRequest, "Pesan minimal 10 karakter")
		return
	}
	subject := in.Category
	if subject == "" {
		subject = "other"
	}
	f, err := h.repo.Create(r.Context(), u.UserID, subject, in.Message)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, f)
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
	if pid := r.URL.Query().Get("proyekId"); pid != "" {
		if v, err := atoi32(pid); err == nil {
			f.ProyekID = &v
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
	out, err := h.repo.ListActive(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}
