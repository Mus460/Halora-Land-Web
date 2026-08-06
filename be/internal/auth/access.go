package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/halora-land/halora-be/internal/models"
)

// RequireAuth responds 401 when no authenticated user is present. This replaces
// the per-route inline 401 check duplicated across all 35 current routes
// (ARCHITECTURE.md §3.9, §7.2 #5).
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if FromContext(r.Context()) == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole responds 403 when the authenticated user's role is not in the
// allowed set. Use for ADMIN-only routes (e.g. /admin/*, ahsp import).
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := FromContext(r.Context())
			if u == nil {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized"})
				return
			}
			if _, ok := allowed[u.Role]; !ok {
				writeJSON(w, http.StatusForbidden, map[string]any{"error": "Forbidden"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AccessLevel expresses the project-team permission tier required for a route.
type AccessLevel int

const (
	AccessView  AccessLevel = iota // owner | team member | admin
	AccessEdit                     // non-viewer team member | owner | admin
	AccessOwner                     // project owner | admin (delete pekerjaan etc.)
)

// ProjectAccess loads a project and verifies the current user meets the given
// access level. Returns the project on success, or sends the appropriate 403/404
// and a nil project on failure. This is the single, central replacement for the
// inline `hasAccess` block repeated across every route today (§2.1, §3.9).
func ProjectAccess(ctx context.Context, pool *pgxpool.Pool, proyekID int32, level AccessLevel) (*models.Proyek, *TimMembership, bool) {
	u := FromContext(ctx)
	if u == nil {
		return nil, nil, false
	}
	var (
		ownerID int32
		timRole *models.RoleTimProyek
	)
	err := pool.QueryRow(ctx, `
		SELECT p."userId", tp.role
		FROM proyek p
		LEFT JOIN tim_proyek tp ON tp."proyekId" = p.id AND tp."userId" = $2
		WHERE p.id = $1 AND p."deletedAt" IS NULL`, proyekID, u.UserID).Scan(&ownerID, &timRole)
	if err != nil {
		return nil, nil, false
	}
	m := &TimMembership{IsOwner: ownerID == u.UserID, IsAdmin: u.Role == models.RoleAdmin, Role: timRole}
	if !m.Has(level) {
		return nil, m, false
	}
	return &models.Proyek{ID: proyekID, UserID: ownerID}, m, true
}

// TimMembership summarizes the current user's relationship to a project.
type TimMembership struct {
	IsOwner bool
	IsAdmin bool
	Role    *models.RoleTimProyek // nil if not a team member
}

func (m *TimMembership) Has(level AccessLevel) bool {
	if m.IsAdmin {
		return true
	}
	switch level {
	case AccessView:
		return m.IsOwner || m.Role != nil
	case AccessEdit:
		if m.IsOwner {
			return true
		}
		return m.Role != nil && *m.Role != models.TimRoleViewer
	case AccessOwner:
		return m.IsOwner
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
