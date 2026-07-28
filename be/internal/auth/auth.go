package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/halora-land/halora-be/internal/models"
)

type ctxKey string

const (
	ctxUser ctxKey = "user"
)

// Verifier validates Supabase access tokens via JWKS (ARCHITECTURE.md §3.3/§7).
type Verifier struct {
	jwksURL     string
	anonKey     string
	projectRef  string
	pool        *pgxpool.Pool
	httpClient  *http.Client

	mu       sync.RWMutex
	keyset   jwt.Keyfunc
	fetchedAt time.Time
	ttl      time.Duration
}

func NewVerifier(pool *pgxpool.Pool, jwksURL, anonKey, projectRef string) *Verifier {
	v := &Verifier{
		jwksURL:    jwksURL,
		anonKey:    anonKey,
		projectRef: projectRef,
		pool:       pool,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		ttl:        15 * time.Minute,
	}
	v.keyset = v.fetchKeyset
	return v
}

// AuthUser is the authenticated principal placed into the request context.
type AuthUser struct {
	UserID      int32
	Email       string
	Role        models.Role
	NamaLengkap string
	AccountType string
	IsDemo      bool
	SupabaseSub string
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

// Authenticate middleware: verifies the bearer/cookie token, loads the local
// user row, auto-links supabaseAuthId when null (ARCHITECTURE.md §2.2 step 4),
// and stores *AuthUser in the context. Calls next with no user on failure so
// downstream handlers/middleware can decide (401 vs public).
func (v *Verifier) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r, v.projectRef)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}
		sub, email, err := v.verifyToken(r.Context(), token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		u, err := v.loadOrCreateUser(r.Context(), sub, email)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
	})
}

func (v *Verifier) verifyToken(ctx context.Context, token string) (sub, email string, err error) {
	parsed, err := jwt.Parse(token, v.keyset)
	if err != nil {
		return "", "", fmt.Errorf("parse jwt: %w", err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}
	if !parsed.Valid {
		return "", "", fmt.Errorf("invalid token")
	}
	if s, ok := claims["sub"].(string); ok {
		sub = s
	}
	if e, ok := claims["email"].(string); ok {
		email = e
	}
	if sub == "" {
		return "", "", fmt.Errorf("missing sub")
	}
	return sub, email, nil
}

// loadOrCreateUser mirrors getCurrentSupabaseUser + login find-or-create
// (ARCHITECTURE.md §2.2 step 4, login/route.ts find-or-create).
func (v *Verifier) loadOrCreateUser(ctx context.Context, sub, email string) (*AuthUser, error) {
	return v.LoadOrCreate(ctx, sub, email)
}

// LoadOrCreate finds or creates the local user row for a Supabase sub/email.
// Exposed so the login handler (which has no cookie yet) can run the same
// find-or-create + auto-link path as the middleware.
func (v *Verifier) LoadOrCreate(ctx context.Context, sub, email string) (*AuthUser, error) {
	tx, err := v.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var (
		u           models.User
		supabaseID  *string
	)
	row := tx.QueryRow(ctx, `
		SELECT id, "namaLengkap", email, role, "accountType", "isDemo", "supabaseAuthId"
		FROM users WHERE "supabaseAuthId" = $1 OR email = $2
		LIMIT 1`, sub, email)
	err = row.Scan(&u.ID, &u.NamaLengkap, &u.Email, &u.Role, &u.AccountType, &u.IsDemo, &supabaseID)
	if err == pgx.ErrNoRows {
		name := email
		if name == "" {
			name = "User"
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO users ("namaLengkap", email, role, "accountType", "isDemo", "supabaseAuthId")
			VALUES ($1, $2, 'USER', 'free', false, $3)`,
			name, email, sub); err != nil {
			return nil, err
		}
		if err := tx.QueryRow(ctx, `
			SELECT id, "namaLengkap", email, role, "accountType", "isDemo", "supabaseAuthId"
			FROM users WHERE "supabaseAuthId" = $1`, sub).
			Scan(&u.ID, &u.NamaLengkap, &u.Email, &u.Role, &u.AccountType, &u.IsDemo, &supabaseID); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if supabaseID == nil {
		if _, err := tx.Exec(ctx, `UPDATE users SET "supabaseAuthId" = $1 WHERE id = $2`, sub, u.ID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &AuthUser{
		UserID:      u.ID,
		Email:       u.Email,
		Role:        u.Role,
		NamaLengkap: u.NamaLengkap,
		AccountType: u.AccountType,
		IsDemo:      u.IsDemo,
		SupabaseSub: sub,
	}, nil
}

func extractToken(r *http.Request, projectRef string) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if c, err := r.Cookie(fmt.Sprintf("sb-%s-auth-token", projectRef)); err == nil && c.Value != "" {
		return tokenFromCookieValue(c.Value)
	}
	return ""
}

// tokenFromCookieValue supports both the raw JWT and the Supabase SSR
// JSON-encoded cookie shape ({access_token: ...}) used by the current FE.
func tokenFromCookieValue(v string) string {
	if strings.HasPrefix(v, "eyJ") {
		return v
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(v), &m); err == nil {
		if at, ok := m["access_token"].(string); ok {
			return at
		}
	}
	return ""
}

func (v *Verifier) fetchKeyset(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}
	v.mu.RLock()
	if v.fetchedAt.IsZero() || time.Since(v.fetchedAt) > v.ttl {
		v.mu.RUnlock()
		v.mu.Lock()
		defer v.mu.Unlock()
		if v.fetchedAt.IsZero() || time.Since(v.fetchedAt) > v.ttl {
			set, err := v.fetchJWKS()
			if err != nil {
				return nil, err
			}
			v.fetchedAt = time.Now()
			v.keyset = set
		}
		ks := v.keyset
		v.mu.RUnlock()
		return ks(t)
	}
	ks := v.keyset
	v.mu.RUnlock()
	return ks(t)
}

func (v *Verifier) fetchJWKS() (jwt.Keyfunc, error) {
	req, _ := http.NewRequest(http.MethodGet, v.jwksURL, nil)
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks status %d", resp.StatusCode)
	}
	set, err := jwt.ParseRSAPublicKeyFromPEM(body)
	if err != nil {
		return nil, fmt.Errorf("parse jwks: %w", err)
	}
	return func(t *jwt.Token) (interface{}, error) { return set, nil }, nil
}
