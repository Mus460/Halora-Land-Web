package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"github.com/halora-land/halora-be/internal/audit"
	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/env"
	"github.com/halora-land/halora-be/internal/models"

	"github.com/halora-land/halora-be/internal/repository"
	"github.com/jackc/pgx/v5"

	"github.com/halora-land/halora-be/service"
)

func newPool(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	m, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	t.Cleanup(func() { m.Close() })
	return m
}

func doReq(t *testing.T, h http.HandlerFunc, method, target, body string, ctx context.Context) *httptest.ResponseRecorder {
	t.Helper()
	var rd *bytes.Reader
	if body == "" {
		rd = bytes.NewReader(nil)
	} else {
		rd = bytes.NewReader([]byte(body))
	}
	r := httptest.NewRequest(method, target, rd)
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if ctx != nil {
		r = r.WithContext(ctx)
	}
	w := httptest.NewRecorder()
	h(w, r)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode %s: %v", w.Body.String(), err)
	}
}

func testConfig() *env.Config {
	return &env.Config{JWTSecret: "sekret-test", IsProd: false}
}

func TestHandlerLoginSuccess(t *testing.T) {
	m := newPool(t)
	hash, _ := auth.HashPassword("rahasia123")
	m.ExpectQuery(`SELECT id, "fullName", email, role, "accountType", "isDemo", "passwordHash" FROM users WHERE email =`).
		WithArgs("admin@haloraland.id").
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo", "passwordHash"}).
			AddRow(int32(1), "Admin", "admin@haloraland.id", models.RoleAdmin, "developer", true, &hash))

	v := auth.NewVerifier(m, "sekret-test")
	h := NewAuthHandler(testConfig(), v, nil)
	w := doReq(t, h.Login, http.MethodPost, "/api/v1/auth/login",
		`{"email":"admin@haloraland.id","password":"rahasia123"}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
	var out struct {
		User struct {
			ID   int32  `json:"id"`
			Role string `json:"role"`
		} `json:"user"`
	}
	decodeBody(t, w, &out)
	if out.User.ID != 1 || out.User.Role != string(models.RoleAdmin) {
		t.Errorf("user = %+v", out.User)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.CookieName || cookies[0].Value == "" {
		t.Fatalf("cookie = %+v", cookies)
	}
	if cookies[0].Secure {
		t.Error("cookie must not be Secure in dev")
	}
}

func TestHandlerLoginValidation(t *testing.T) {
	h := NewAuthHandler(testConfig(), auth.NewVerifier(newPool(t), "x"), nil)
	w := doReq(t, h.Login, http.MethodPost, "/api/v1/auth/login", `{"email":"","password":""}`, nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d", w.Code)
	}
	w = doReq(t, h.Login, http.MethodPost, "/api/v1/auth/login", `{bad json`, nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("bad json status = %d", w.Code)
	}
}

func TestHandlerLoginWrongPassword(t *testing.T) {
	m := newPool(t)
	hash, _ := auth.HashPassword("benar123")
	m.ExpectQuery(`SELECT id, "fullName", email, role, "accountType", "isDemo", "passwordHash" FROM users WHERE email =`).
		WithArgs("admin@haloraland.id").
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo", "passwordHash"}).
			AddRow(int32(1), "Admin", "admin@haloraland.id", models.RoleAdmin, "developer", true, &hash))
	v := auth.NewVerifier(m, "sekret-test")
	h := NewAuthHandler(testConfig(), v, nil)
	w := doReq(t, h.Login, http.MethodPost, "/api/v1/auth/login",
		`{"email":"admin@haloraland.id","password":"salah"}`, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d body %s", w.Code, w.Body.String())
	}
}

func TestHandlerRegisterRequiresAdmin(t *testing.T) {
	m := newPool(t)
	h := NewAuthHandler(testConfig(), auth.NewVerifier(m, "x"), nil)
	w := doReq(t, h.Register, http.MethodPost, "/api/v1/auth/register",
		`{"email":"a@b.id","password":"123456"}`, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d want 403", w.Code)
	}
}

func TestHandlerRegisterAdminOk(t *testing.T) {
	m := newPool(t)
	actor := &auth.AuthUser{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(context.Background(), actor)
	m.ExpectQuery(`INSERT INTO users`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo"}).
			AddRow(int32(2), "admin", "admin@haloraland.id", models.RoleAdmin, "developer", true))

	h := NewAuthHandler(testConfig(), auth.NewVerifier(m, "x"), nil)
	w := doReq(t, h.Register, http.MethodPost, "/api/v1/auth/register",
		`{"email":"admin@haloraland.id","password":"123456"}`, ctx)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
	var out struct {
		User struct {
			ID int32 `json:"id"`
		} `json:"user"`
	}
	decodeBody(t, w, &out)
	if out.User.ID != 2 {
		t.Errorf("user = %+v", out.User)
	}
}

func TestHandlerRegisterPasswordTooShort(t *testing.T) {
	m := newPool(t)
	actor := &auth.AuthUser{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(context.Background(), actor)
	h := NewAuthHandler(testConfig(), auth.NewVerifier(m, "x"), nil)
	w := doReq(t, h.Register, http.MethodPost, "/api/v1/auth/register",
		`{"email":"a@b.id","password":"123"}`, ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d", w.Code)
	}
}

func TestHandlerMe(t *testing.T) {
	h := NewAuthHandler(testConfig(), auth.NewVerifier(newPool(t), "x"), nil)
	u := &auth.AuthUser{UserID: 3, FullName: "Budi", Email: "b@b.id", Role: models.RoleUser}
	w := doReq(t, h.Me, http.MethodGet, "/api/v1/auth/me", "", auth.WithUser(context.Background(), u))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var out struct {
		User struct {
			ID   int32  `json:"id"`
			Name string `json:"fullName"`
		} `json:"user"`
	}
	decodeBody(t, w, &out)
	if out.User.ID != 3 || out.User.Name != "Budi" {
		t.Errorf("user = %+v", out.User)
	}
	w = doReq(t, h.Me, http.MethodGet, "/api/v1/auth/me", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("anonymous status = %d", w.Code)
	}
}

func TestHandlerLogoutClearsCookie(t *testing.T) {
	h := NewAuthHandler(testConfig(), auth.NewVerifier(newPool(t), "x"), nil)
	w := doReq(t, h.Logout, http.MethodPost, "/api/v1/auth/logout", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.CookieName || cookies[0].MaxAge != -1 {
		t.Errorf("cookies = %+v", cookies)
	}
}

func TestHandlerFeedbackCreateAndList(t *testing.T) {
	m := newPool(t)
	u := &auth.AuthUser{UserID: 1, Role: models.RoleUser}
	m.ExpectQuery(`INSERT INTO feedback`).WithArgs(int32(1), "Saran", "tambah fitur dong").
		WillReturnRows(pgxmock.NewRows([]string{"id", "userId", "subject", "message", "status", "createdAt", "updatedAt"}).
			AddRow(int32(1), int32(1), "Saran", "tambah fitur dong", models.FeedbackOpen, time.Now(), time.Now()))
	m.ExpectQuery(`FROM feedback WHERE "userId"`).WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "userId", "subject", "message", "status", "createdAt", "updatedAt"}).
			AddRow(int32(1), int32(1), "Saran", "tambah fitur dong", models.FeedbackOpen, time.Now(), time.Now()))

	h := NewFeedbackHandler(repository.NewFeedbackRepo(m))
	ctx := auth.WithUser(context.Background(), u)
	w := doReq(t, h.Create, http.MethodPost, "/api/v1/feedback",
		`{"subject":"Saran","message":"tambah fitur dong"}`, ctx)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d body %s", w.Code, w.Body.String())
	}
	w = doReq(t, h.Create, http.MethodPost, "/api/v1/feedback", `{"message":"pendek"}`, ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("short message status = %d", w.Code)
	}
	w = doReq(t, h.List, http.MethodGet, "/api/v1/feedback", "", ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d", w.Code)
	}
	var out struct {
		Feedback []struct {
			Subject string `json:"subject"`
		} `json:"feedback"`
	}
	decodeBody(t, w, &out)
	if len(out.Feedback) != 1 || out.Feedback[0].Subject != "Saran" {
		t.Errorf("feedback = %+v", out.Feedback)
	}
}

func TestHandlerMonitoringRequiresAccess(t *testing.T) {
	m := newPool(t)
	h := NewMonitoringHandler(m, service.NewProgressService(m))
	w := doReq(t, h.List, http.MethodGet, "/api/v1/monitoring", "", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing projectId status = %d", w.Code)
	}
}

func TestHandlerMapPgErr(t *testing.T) {
	if st, _ := mapPgErr(nil); st != http.StatusOK {
		t.Errorf("nil = %d", st)
	}
	if st, _ := mapPgErr(pgx.ErrNoRows); st != http.StatusNotFound {
		t.Errorf("no rows = %d", st)
	}
	if st, _ := mapPgErr(errors.New("ERROR: duplicate key value violates unique constraint \"x\" (SQLSTATE 23505)")); st != http.StatusConflict {
		t.Errorf("23505 = %d", st)
	}
	if st, _ := mapPgErr(errors.New("boom")); st != http.StatusInternalServerError {
		t.Errorf("generic = %d", st)
	}
}

func TestHandlerAuditLoggerWired(t *testing.T) {
	m := newPool(t)
	hash, _ := auth.HashPassword("rahasia123")
	m.ExpectQuery(`SELECT id, "fullName", email, role, "accountType", "isDemo", "passwordHash" FROM users WHERE email =`).
		WithArgs("admin@haloraland.id").
		WillReturnRows(pgxmock.NewRows([]string{"id", "fullName", "email", "role", "accountType", "isDemo", "passwordHash"}).
			AddRow(int32(1), "Admin", "admin@haloraland.id", models.RoleAdmin, "developer", true, &hash))
	// async audit write
	m.ExpectExec(`INSERT INTO audit_log`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	al := audit.New(m, 10)
	defer al.Close()
	h := NewAuthHandler(testConfig(), auth.NewVerifier(m, "sekret-test"), al)
	w := doReq(t, h.Login, http.MethodPost, "/api/v1/auth/login",
		`{"email":"admin@haloraland.id","password":"rahasia123"}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	al.Close()
	if err := m.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
