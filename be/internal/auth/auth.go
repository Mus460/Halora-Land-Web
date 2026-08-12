package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/halora-land/halora-be/internal/database"

	"github.com/halora-land/halora-be/internal/models"
)

type ctxKey string

const (
	ctxUser ctxKey = "user"
)

// Verifier authenticates users against the local DB using a signed session
// JWT issued at login (HS256, golang-jwt). No external auth provider is used:
// the token is verified with the shared JWT secret and the user row is loaded
// from Postgres on every request.
type Verifier struct {
	pool      database.Pool
	jwtSecret string
}

// NewVerifier builds a local-auth verifier.
func NewVerifier(pool database.Pool, jwtSecret string) *Verifier {
	return &Verifier{pool: pool, jwtSecret: jwtSecret}
}

// AuthUser is the authenticated principal placed into the request context.
type AuthUser struct {
	UserID      int32
	Email       string
	Role        models.Role
	FullName    string
	AccountType string
	IsDemo      bool
}

// FromContext extracts the authenticated user from the context, or nil.
func FromContext(ctx context.Context) *AuthUser {
	v, _ := ctx.Value(ctxUser).(*AuthUser)
	return v
}

// WithUser returns a context carrying the authenticated user.
func WithUser(ctx context.Context, u *AuthUser) context.Context {
	return context.WithValue(ctx, ctxUser, u)
}

// Authenticate middleware: verifies the session token from the Authorization
// header or the session cookie, loads the local user row, and stores *AuthUser
// in the context. Calls next with no user on failure so downstream
// handlers/middleware can decide (401 vs public).
func (v *Verifier) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}
		userID, err := VerifySessionToken(token, v.jwtSecret)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		u, err := v.loadUser(r.Context(), userID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
	})
}

// loadUser loads the user row for the given ID.
func (v *Verifier) loadUser(ctx context.Context, userID int32) (*AuthUser, error) {
	var u models.User
	err := v.pool.QueryRow(ctx, `
		SELECT id, "fullName", email, role, "accountType", "isDemo"
		FROM users WHERE id = $1`, userID).
		Scan(&u.ID, &u.FullName, &u.Email, &u.Role, &u.AccountType, &u.IsDemo)
	if err != nil {
		return nil, err
	}
	return &AuthUser{
		UserID:      u.ID,
		Email:       u.Email,
		Role:        u.Role,
		FullName:    u.FullName,
		AccountType: u.AccountType,
		IsDemo:      u.IsDemo,
	}, nil
}

// DefaultAdminEmail is the bootstrap admin created when the users table is empty.
const DefaultAdminEmail = "admin@haloraland.id"

// EnsureDefaultAdmin creates the bootstrap admin account (admin@haloraland.id /
// admin123) only when the users table is empty, so a fresh database has a way
// in. Existing databases are never touched. Returns true when a new admin was
// created.
func (v *Verifier) EnsureDefaultAdmin(ctx context.Context) (bool, error) {
	var count int
	if err := v.pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count); err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	hash, err := HashPassword("admin123")
	if err != nil {
		return false, err
	}
	_, err = v.pool.Exec(ctx, `
		INSERT INTO users ("fullName", email, role, "accountType", "isDemo", "passwordHash")
		VALUES ('Admin Halora', $1, 'ADMIN', 'free', false, $2)`,
		DefaultAdminEmail, hash)
	return err == nil, err
}

// CookieName is the HTTP-only session cookie set at login.
const CookieName = "halora_session"

func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if c, err := r.Cookie(CookieName); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}
