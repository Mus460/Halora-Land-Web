package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/halora-land/halora-be/internal/ahsp"
	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/handler"
	"github.com/halora-land/halora-be/internal/models"
	"github.com/halora-land/halora-be/internal/ratelimit"
	"github.com/halora-land/halora-be/internal/repository"
	"github.com/halora-land/halora-be/service"
)

const window15 = 15 * time.Minute

func (app *App) routes() http.Handler {
	// Repositories
	proyekRepo := repository.NewProjectRepo(app.pool)
	pekerjaanRepo := repository.NewWorkItemRepo(app.pool)
	maRepo := repository.NewAnalysisMasterRepo(app.pool)
	mhRepo := repository.NewPriceMasterRepo(app.pool)
	clientRepo := repository.NewClientRepo(app.pool)
	dashRepo := repository.NewDashboardRepo(app.pool)
	auditRepo := repository.NewAuditLogRepo(app.pool)
	feedbackRepo := repository.NewFeedbackRepo(app.pool)

	// Services
	snap := service.NewSnapshotService(app.pool, pekerjaanRepo, maRepo, app.audit)
	rab := service.NewRABService(app.pool, pekerjaanRepo, app.cfg.PPNRate)
	progress := service.NewProgressService(app.pool)
	importer := ahsp.NewImporter(app.pool)

	// Handlers
	authH := handler.NewAuthHandler(app.cfg, app.verifier, app.audit)
	proyekH := handler.NewProjectHandler(app.pool, proyekRepo, rab)
	pekerjaanH := handler.NewWorkItemHandler(app.pool, pekerjaanRepo, snap, progress, rab)
	subH := handler.NewProjectSubHandler(app.pool, proyekRepo, rab, snap, progress)
	maH := handler.NewAnalysisMasterHandler(app.pool, maRepo)
	mhH := handler.NewPriceMasterHandler(app.pool, mhRepo)
	clientH := handler.NewClientHandler(clientRepo)
	dashH := handler.NewDashboardHandler(app.pool, dashRepo)
	feedbackH := handler.NewFeedbackHandler(feedbackRepo)
	auditH := handler.NewAuditLogHandler(auditRepo)
	adminH := handler.NewAdminAHSPHandler(app.pool, importer, app.ahspPath)
	monH := handler.NewMonitoringHandler(app.pool, progress)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(app.verifier.Authenticate)

	r.Route("/api/v1", func(r chi.Router) {
		// Public auth (rate-limited, ARCHITECTURE.md §3.8)
		r.With(rateLimit(app.limiter, "login", 10, ratelimit.ByIP("login"))).
			Post("/auth/login", authH.Login)
		r.With(rateLimit(app.limiter, "resend", 3, ratelimit.ByIP("resend"))).
			Post("/auth/resend-verification", authH.ResendVerification)

		// Authenticated
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth)
			r.Post("/auth/logout", authH.Logout)
			r.Get("/auth/me", authH.Me)
			r.Put("/auth/me", authH.Me)
			r.With(rateLimit(app.limiter, "pwupdate", 5, ratelimit.ByIP("pwupdate"))).
				Post("/auth/update-password", authH.UpdatePassword)

			r.Get("/projects", proyekH.List)
			r.Post("/projects", proyekH.Create)
			r.Post("/projects/import", proyekH.Import)
			r.Get("/projects/{id}", proyekH.Get)
			r.Put("/projects/{id}", proyekH.Update)
			r.Delete("/projects/{id}", proyekH.Delete)
			r.Get("/projects/{id}/recaps", subH.RecapGet)
			r.Post("/projects/{id}/recalculate-all", subH.RecalculateAll)
			r.Get("/projects/{id}/transactions", subH.TransactionList)
			r.Post("/projects/{id}/transactions", subH.TransactionCreate)
			r.Put("/projects/{id}/transactions/{transactionId}", subH.TransactionApprove)
			r.Delete("/projects/{id}/transactions/{transactionId}", subH.TransactionDelete)
			r.Get("/projects/{id}/logistics", subH.LogisticsList)
			r.Post("/projects/{id}/logistics", subH.LogisticsCreate)
			r.Get("/projects/{id}/invoices", subH.InvoiceList)
			r.Post("/projects/{id}/invoices", subH.InvoiceCreate)
			r.Put("/projects/{id}/invoices/{invoiceId}", subH.InvoiceUpdate)
			r.Get("/projects/{id}/s-curve", subH.SCurve)
			r.Post("/projects/{id}/work_items/from-ahsp", pekerjaanH.FromAHSP)

			r.Get("/work-items", pekerjaanH.List)
			r.Post("/work-items", pekerjaanH.Create)
			r.Get("/work-items/{id}", pekerjaanH.Get)
			r.Put("/work-items/{id}", pekerjaanH.Update)
			r.Delete("/work-items/{id}", pekerjaanH.Delete)
			r.Get("/work-items/{id}/analysis", pekerjaanH.ListAnalisa)
			r.Put("/work-items/{id}/details", pekerjaanH.UpdateDetails)
			r.Post("/work-items/{id}/recalculate", pekerjaanH.Recalculate)
			r.Put("/work-items/{id}/progress", pekerjaanH.UpdateProgress)
			r.Get("/work-items/{id}/progress-logs", pekerjaanH.ProgressLogs)
			r.Get("/work-items/{id}/validate-snapshot", pekerjaanH.ValidateSnapshot)

			r.Get("/analysis-masters", maH.List)
			r.Post("/analysis-masters", maH.Create)
			r.Get("/analysis-masters/search", maH.Search)
			r.Get("/analysis-masters/{id}", maH.Get)
			r.Put("/analysis-masters/{id}", maH.Update)
			r.Post("/analysis-masters/{id}/copy", maH.Copy)
			r.Delete("/analysis-masters/{id}", maH.Delete)
			r.Get("/analysis-masters/{id}/components", maH.ListComponents)
			r.Post("/analysis-masters/{id}/components", maH.CreateComponent)
			r.Put("/analysis-masters/{id}/components/{componentId}", maH.UpdateComponent)
			r.Delete("/analysis-masters/{id}/components/{componentId}", maH.DeleteComponent)

			r.Get("/price-masters", mhH.List)
			r.Post("/price-masters", mhH.Create)
			r.Get("/price-masters/{id}", mhH.Get)
			r.Put("/price-masters/{id}", mhH.Update)
			r.Delete("/price-masters/{id}", mhH.Delete)

			r.Get("/clients", clientH.List)
			r.Post("/clients", clientH.Create)
			r.Put("/clients/{id}", clientH.Update)
			r.Delete("/clients/{id}", clientH.Delete)

			r.Get("/dashboard/stats", dashH.Stats)
			r.Get("/audit-log", auditH.List)
			r.Get("/feedback", feedbackH.List)
			r.Post("/feedback", feedbackH.Create)
			r.Get("/monitoring", monH.List)

			// Admin-only (ARCHITECTURE.md §3.9 — protect at middleware layer)
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(models.RoleAdmin))
				r.Post("/auth/register", authH.Register)
				r.Get("/users", authH.ListUsers)
				r.Delete("/users/{id}", authH.DeleteUser)
				r.Get("/admin/ahsp/import", adminH.ImportStatus)
				r.Post("/admin/ahsp/import", adminH.Import)
			})
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})
	return r
}

func rateLimit(l *ratelimit.Limiter, route string, limit int, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return l.Middleware(keyFn, limit, window15, next)
	}
}
