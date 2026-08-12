package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/halora-land/halora-be/internal/models"
)

// --- password hashing ---

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "s3cret" || hash == "" {
		t.Fatal("hash looks wrong")
	}
	if err := CheckPassword(hash, "s3cret"); err != nil {
		t.Errorf("CheckPassword correct: %v", err)
	}
	if err := CheckPassword(hash, "wrong"); err == nil {
		t.Error("CheckPassword wrong password should fail")
	}
	if err := CheckPassword("not-a-hash", "x"); err == nil {
		t.Error("CheckPassword garbage hash should fail")
	}
}

// --- session JWT ---

func TestIssueAndVerifySessionToken(t *testing.T) {
	tok, err := IssueSessionToken(42, "sekret")
	if err != nil {
		t.Fatalf("IssueSessionToken: %v", err)
	}
	id, err := VerifySessionToken(tok, "sekret")
	if err != nil {
		t.Fatalf("VerifySessionToken: %v", err)
	}
	if id != 42 {
		t.Errorf("userID = %d want 42", id)
	}
}

func TestVerifySessionTokenWrongSecret(t *testing.T) {
	tok, _ := IssueSessionToken(1, "sekret")
	if _, err := VerifySessionToken(tok, "other"); err == nil {
		t.Fatal("expected error with wrong secret")
	}
}

func TestVerifySessionTokenEmptySecret(t *testing.T) {
	if _, err := IssueSessionToken(1, ""); err == nil {
		t.Fatal("IssueSessionToken with empty secret should fail")
	}
	if _, err := VerifySessionToken("abc", ""); err == nil {
		t.Fatal("VerifySessionToken with empty secret should fail")
	}
}

func TestVerifySessionTokenGarbage(t *testing.T) {
	if _, err := VerifySessionToken("not.a.token", "sekret"); err == nil {
		t.Fatal("expected error for garbage token")
	}
}

func TestVerifySessionTokenExpired(t *testing.T) {
	claims := SessionClaims{
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("sekret"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := VerifySessionToken(tok, "sekret"); err == nil {
		t.Fatal("expected error for expired token")
	}
}

// --- context helpers ---

func TestWithUserFromContext(t *testing.T) {
	u := &AuthUser{UserID: 7, Email: "a@b.c", Role: models.RoleAdmin}
	ctx := WithUser(context.Background(), u)
	if got := FromContext(ctx); got != u {
		t.Errorf("FromContext = %v want same pointer", got)
	}
	if FromContext(context.Background()) != nil {
		t.Error("FromContext on empty context should be nil")
	}
}

// --- token extraction ---

func TestExtractTokenBearer(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer tok123")
	if got := extractToken(r); got != "tok123" {
		t.Errorf("got %q", got)
	}
}

func TestExtractTokenCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: CookieName, Value: "cookieval"})
	if got := extractToken(r); got != "cookieval" {
		t.Errorf("got %q", got)
	}
}

func TestExtractTokenBearerWinsOverCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer tok123")
	r.AddCookie(&http.Cookie{Name: CookieName, Value: "cookieval"})
	if got := extractToken(r); got != "tok123" {
		t.Errorf("got %q", got)
	}
}

func TestExtractTokenNone(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	if got := extractToken(r); got != "" {
		t.Errorf("got %q", got)
	}
}

// --- Authenticate middleware (DB-backed) ---

const userSelect = `SELECT id, "fullName", email, role, "accountType", "isDemo" FROM users WHERE id =`

func newMock(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	m, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	t.Cleanup(func() { m.Close() })
	return m
}

func TestAuthenticateWithValidToken(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	tok, _ := IssueSessionToken(5, "sekret")

	m.ExpectQuery(userSelect).
		WithArgs(int32(5)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo"}).
			AddRow(int32(5), "Budi", "budi@x.id", models.RoleAdmin, "free", false))

	var got *AuthUser
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = FromContext(r.Context())
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	v.Authenticate(next).ServeHTTP(httptest.NewRecorder(), req)

	if got == nil {
		t.Fatal("expected authenticated user in context")
	}
	if got.UserID != 5 || got.Email != "budi@x.id" || got.Role != models.RoleAdmin {
		t.Errorf("user = %+v", got)
	}
}

func TestAuthenticateNoToken(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if FromContext(r.Context()) != nil {
			t.Error("no user expected")
		}
	})
	v.Authenticate(next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Fatal("next should be called")
	}
}

func TestAuthenticateInvalidToken(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	var got *AuthUser
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = FromContext(r.Context()) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer bad.token.here")
	v.Authenticate(next).ServeHTTP(httptest.NewRecorder(), req)
	if got != nil {
		t.Error("expected no user for invalid token")
	}
}

func TestAuthenticateDBError(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	tok, _ := IssueSessionToken(5, "sekret")
	m.ExpectQuery(userSelect).WithArgs(int32(5)).WillReturnError(pgx.ErrNoRows)
	var got *AuthUser
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = FromContext(r.Context()) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	v.Authenticate(next).ServeHTTP(httptest.NewRecorder(), req)
	if got != nil {
		t.Error("expected no user when DB lookup fails")
	}
}

// --- LocalLogin / LocalRegister / EnsureDefaultAdmin ---

func TestLocalLoginSuccess(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	hashB, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	hash := string(hashB)

	m.ExpectQuery(`SELECT id, "fullName", email, role, "accountType", "isDemo", "passwordHash" FROM users WHERE email =`).
		WithArgs("budi@x.id").
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo", "passwordHash"}).
			AddRow(int32(3), "Budi", "budi@x.id", models.RoleAdmin, "free", false, &hash))

	u, err := v.LocalLogin(context.Background(), "budi@x.id", "pass123")
	if err != nil {
		t.Fatalf("LocalLogin: %v", err)
	}
	if u.UserID != 3 || u.FullName != "Budi" {
		t.Errorf("user = %+v", u)
	}
}

func TestLocalLoginWrongPassword(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	hashB, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	hash := string(hashB)

	m.ExpectQuery(`SELECT id, "fullName", email, role, "accountType", "isDemo", "passwordHash" FROM users WHERE email =`).
		WithArgs("budi@x.id").
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo", "passwordHash"}).
			AddRow(int32(3), "Budi", "budi@x.id", models.RoleAdmin, "free", false, &hash))

	if _, err := v.LocalLogin(context.Background(), "budi@x.id", "nope"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLocalLoginUnknownEmail(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	m.ExpectQuery(`SELECT id, "fullName", email, role, "accountType", "isDemo", "passwordHash" FROM users WHERE email =`).
		WithArgs("ghost@x.id").
		WillReturnError(pgx.ErrNoRows)
	if _, err := v.LocalLogin(context.Background(), "ghost@x.id", "x"); err == nil {
		t.Fatal("expected error for unknown email")
	}
}

func TestLocalLoginNoPasswordHash(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	m.ExpectQuery(`SELECT id, "fullName", email, role, "accountType", "isDemo", "passwordHash" FROM users WHERE email =`).
		WithArgs("oauth@x.id").
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo", "passwordHash"}).
			AddRow(int32(3), "Oauth", "oauth@x.id", models.RoleUser, "free", false, nil))
	if _, err := v.LocalLogin(context.Background(), "oauth@x.id", "x"); err == nil {
		t.Fatal("expected error for account without password")
	}
}

func TestLocalRegister(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	m.ExpectQuery(`INSERT INTO users`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo"}).
			AddRow(int32(9), "Sari", "sari@x.id", models.RoleAdmin, "free", false))
	u, err := v.LocalRegister(context.Background(), "Sari", "sari@x.id", "pw12345")
	if err != nil {
		t.Fatalf("LocalRegister: %v", err)
	}
	if u.UserID != 9 || u.Email != "sari@x.id" {
		t.Errorf("user = %+v", u)
	}
}

func TestEnsureDefaultAdminCreatesWhenEmpty(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	m.ExpectQuery(`SELECT count\(\*\) FROM users`).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	m.ExpectExec(`INSERT INTO users`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	created, err := v.EnsureDefaultAdmin(context.Background())
	if err != nil {
		t.Fatalf("EnsureDefaultAdmin: %v", err)
	}
	if !created {
		t.Error("expected created=true on empty table")
	}
}

func TestEnsureDefaultAdminSkipsWhenNotEmpty(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	m.ExpectQuery(`SELECT count\(\*\) FROM users`).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(3))
	created, err := v.EnsureDefaultAdmin(context.Background())
	if err != nil {
		t.Fatalf("EnsureDefaultAdmin: %v", err)
	}
	if created {
		t.Error("expected created=false when users exist")
	}
}

func TestEnsureDefaultAdminDBError(t *testing.T) {
	m := newMock(t)
	v := NewVerifier(m, "sekret")
	m.ExpectQuery(`SELECT count\(\*\) FROM users`).WillReturnError(errors.New("boom"))
	if _, err := v.EnsureDefaultAdmin(context.Background()); err == nil {
		t.Fatal("expected error propagation")
	}
}

// --- RequireAuth / RequireRole middleware ---

func TestRequireAuthWithoutUser(t *testing.T) {
	rec := httptest.NewRecorder()
	RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not run")
	})).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("code = %d want 401", rec.Code)
	}
}

func TestRequireAuthWithUser(t *testing.T) {
	u := &AuthUser{UserID: 1, Role: models.RoleAdmin}
	req := httptest.NewRequest("GET", "/", nil).WithContext(WithUser(reqCtx(), u))
	rec := httptest.NewRecorder()
	ran := false
	RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ran = true })).ServeHTTP(rec, req)
	if !ran {
		t.Error("next should run with authenticated user")
	}
}

func TestRequireRoleWithoutUser(t *testing.T) {
	rec := httptest.NewRecorder()
	RequireRole(models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not run")
	})).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("code = %d want 401", rec.Code)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	u := &AuthUser{UserID: 1, Role: models.RoleUser}
	req := httptest.NewRequest("GET", "/", nil).WithContext(WithUser(reqCtx(), u))
	rec := httptest.NewRecorder()
	RequireRole(models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not run")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("code = %d want 403", rec.Code)
	}
}

func TestRequireRoleAllowed(t *testing.T) {
	u := &AuthUser{UserID: 1, Role: models.RoleAdmin}
	req := httptest.NewRequest("GET", "/", nil).WithContext(WithUser(reqCtx(), u))
	ran := false
	RequireRole(models.RoleAdmin, models.RoleUser)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})).ServeHTTP(httptest.NewRecorder(), req)
	if !ran {
		t.Error("next should run for allowed role")
	}
}

func reqCtx() context.Context { return context.Background() }

// --- TeamMembership.Has matrix ---

func teamRole(r models.TeamRole) *models.TeamRole { return &r }

func TestTeamMembershipHas(t *testing.T) {
	viewer := teamRole(models.TeamRoleViewer)
	editor := teamRole(models.TeamRoleEditor)

	cases := []struct {
		name  string
		m     TeamMembership
		level AccessLevel
		want  bool
	}{
		{"owner view", TeamMembership{IsOwner: true}, AccessView, true},
		{"owner edit", TeamMembership{IsOwner: true}, AccessEdit, true},
		{"owner owner", TeamMembership{IsOwner: true}, AccessOwner, true},
		{"admin all", TeamMembership{IsAdmin: true}, AccessView, true},
		{"admin all2", TeamMembership{IsAdmin: true}, AccessOwner, true},
		{"viewer view", TeamMembership{Role: viewer}, AccessView, true},
		{"viewer edit", TeamMembership{Role: viewer}, AccessEdit, false},
		{"viewer owner", TeamMembership{Role: viewer}, AccessOwner, false},
		{"editor view", TeamMembership{Role: editor}, AccessView, true},
		{"editor edit", TeamMembership{Role: editor}, AccessEdit, true},
		{"editor owner", TeamMembership{Role: editor}, AccessOwner, false},
		{"stranger view", TeamMembership{}, AccessView, false},
		{"stranger edit", TeamMembership{}, AccessEdit, false},
		{"stranger owner", TeamMembership{}, AccessOwner, false},
	}
	for _, c := range cases {
		if got := c.m.Has(c.level); got != c.want {
			t.Errorf("%s: Has(%v) = %v want %v", c.name, c.level, got, c.want)
		}
	}
}
