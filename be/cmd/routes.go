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
	proyekRepo := repository.NewProyekRepo(app.pool)
	pekerjaanRepo := repository.NewPekerjaanRepo(app.pool)
	maRepo := repository.NewMasterAnalisaRepo(app.pool)
	mhRepo := repository.NewMasterHargaRepo(app.pool)
	rekapRepo := repository.NewRekapRepo(app.pool)
	dashRepo := repository.NewDashboardRepo(app.pool)
	auditRepo := repository.NewAuditLogRepo(app.pool)
	feedbackRepo := repository.NewFeedbackRepo(app.pool)
	newsRepo := repository.NewNewsRepo(app.pool)

	// Services
	snap := service.NewSnapshotService(app.pool, pekerjaanRepo, maRepo, app.audit)
	rab := service.NewRABService(app.pool, pekerjaanRepo, rekapRepo, app.cfg.OverheadRate, app.cfg.PPNRate)
	importer := ahsp.NewImporter(app.pool)

	// Handlers
	authH := handler.NewAuthHandler(app.cfg, app.verifier, app.audit)
	proyekH := handler.NewProyekHandler(app.pool, proyekRepo)
	pekerjaanH := handler.NewPekerjaanHandler(app.pool, pekerjaanRepo, snap)
	subH := handler.NewProyekSubHandler(app.pool, proyekRepo, rekapRepo, rab, snap)
	maH := handler.NewMasterAnalisaHandler(app.pool, maRepo)
	mhH := handler.NewMasterHargaHandler(app.pool, mhRepo)
	dashH := handler.NewDashboardHandler(app.pool, dashRepo)
	feedbackH := handler.NewFeedbackHandler(feedbackRepo)
	auditH := handler.NewAuditLogHandler(auditRepo)
	newsH := handler.NewNewsHandler(newsRepo)
	adminH := handler.NewAdminAHSPHandler(app.pool, importer, app.ahspPath)
	monH := handler.NewMonitoringHandler(app.pool)

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
		r.With(rateLimit(app.limiter, "register", 5, ratelimit.ByIP("register"))).
			Post("/auth/register", authH.Register)
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

			r.Get("/proyek", proyekH.List)
			r.Post("/proyek", proyekH.Create)
			r.Get("/proyek/{id}", proyekH.Get)
			r.Put("/proyek/{id}", proyekH.Update)
			r.Delete("/proyek/{id}", proyekH.Delete)
			r.Get("/proyek/{id}/rekap", subH.RekapGet)
			r.Put("/proyek/{id}/rekap", subH.RekapPut)
			r.Post("/proyek/{id}/recalculate-all", subH.RecalculateAll)
			r.Get("/proyek/{id}/realisasi", subH.RealisasiList)
			r.Get("/proyek/{id}/logistik", subH.LogistikList)
			r.Get("/proyek/{id}/invoice", subH.InvoiceList)
			r.Get("/proyek/{id}/kurva-s", subH.KurvaS)
			r.Post("/proyek/{id}/pekerjaan/from-ahsp", pekerjaanH.FromAHSP)

			r.Get("/pekerjaan", pekerjaanH.List)
			r.Post("/pekerjaan", pekerjaanH.Create)
			r.Get("/pekerjaan/{id}", pekerjaanH.Get)
			r.Put("/pekerjaan/{id}", pekerjaanH.Update)
			r.Delete("/pekerjaan/{id}", pekerjaanH.Delete)
			r.Get("/pekerjaan/{id}/analisa", pekerjaanH.ListAnalisa)
			r.Post("/pekerjaan/{id}/recalculate", pekerjaanH.Recalculate)
			r.Get("/pekerjaan/{id}/validate-snapshot", pekerjaanH.ValidateSnapshot)

			r.Get("/master-analisa", maH.List)
			r.Post("/master-analisa", maH.Create)
			r.Get("/master-analisa/search", maH.Search)
			r.Get("/master-analisa/{id}", maH.Get)
			r.Delete("/master-analisa/{id}", maH.Delete)
			r.Get("/master-analisa/{id}/rincian", maH.ListRincian)
			r.Post("/master-analisa/{id}/rincian", maH.CreateRincian)
			r.Delete("/master-analisa/{id}/rincian/{rincianId}", maH.DeleteRincian)

			r.Get("/master-harga", mhH.List)
			r.Post("/master-harga", mhH.Create)
			r.Get("/master-harga/{id}", mhH.Get)
			r.Put("/master-harga/{id}", mhH.Update)
			r.Delete("/master-harga/{id}", mhH.Delete)

			r.Get("/dashboard/stats", dashH.Stats)
			r.Get("/audit-log", auditH.List)
			r.Get("/feedback", feedbackH.List)
			r.Post("/feedback", feedbackH.Create)
			r.Get("/news", newsH.List)
			r.Get("/monitoring", monH.List)

			// Admin-only (ARCHITECTURE.md §3.9 — protect at middleware layer)
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(models.RoleAdmin))
				r.Get("/admin/ahsp/import", adminH.ImportStatus)
				r.Post("/admin/ahsp/import", adminH.Import)
				r.Post("/news", newsH.Create)
				r.Put("/news/{id}", newsH.Update)
				r.Delete("/news/{id}", newsH.Delete)
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
