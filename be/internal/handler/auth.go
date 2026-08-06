package handler

import (
	"net/http"
	"strings"

	"github.com/halora-land/halora-be/internal/audit"
	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/env"
	"github.com/halora-land/halora-be/internal/models"
)

// AuthHandler implements local DB auth (bcrypt password hashing + signed
// session JWT cookie). No external auth provider is involved.
type AuthHandler struct {
	cfg      *env.Config
	verifier *auth.Verifier
	audit    *audit.Logger
}

func NewAuthHandler(cfg *env.Config, v *auth.Verifier, al *audit.Logger) *AuthHandler {
	return &AuthHandler{cfg: cfg, verifier: v, audit: al}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Email == "" || in.Password == "" {
		writeError(w, http.StatusBadRequest, "email dan password wajib diisi")
		return
	}
	u, err := h.verifier.LocalLogin(r.Context(), in.Email, in.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	token, err := auth.IssueSessionToken(u.UserID, h.cfg.JWTSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal membuat sesi")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: auth.CookieName, Value: token, Path: "/",
		MaxAge: 86400, HttpOnly: true, Secure: h.cfg.IsProd,
		SameSite: http.SameSiteStrictMode,
	})
	if h.audit != nil {
		p := audit.Params{Action: audit.ActionLogin, EntityType: "USER", EntityID: &u.UserID, UserID: u.UserID}
		p.FromRequest(r)
		h.audit.Log(p)
	}
	h.writeUser(w, http.StatusOK, u)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Registration is admin-only: only an authenticated ADMIN may create users.
	actor := auth.FromContext(r.Context())
	if actor == nil || actor.Role != models.RoleAdmin {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var in struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		NamaLengkap string `json:"namaLengkap"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Email == "" || in.Password == "" {
		writeError(w, http.StatusBadRequest, "email dan password wajib diisi")
		return
	}
	if len(in.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Password minimal 6 karakter")
		return
	}
	if in.NamaLengkap == "" {
		in.NamaLengkap = strings.SplitN(in.Email, "@", 2)[0]
	}
	u, err := h.verifier.LocalRegister(r.Context(), in.NamaLengkap, in.Email, in.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.audit != nil {
		p := audit.Params{Action: audit.ActionRegister, EntityType: "USER", EntityID: &u.UserID, UserID: u.UserID}
		p.FromRequest(r)
		h.audit.Log(p)
	}
	h.writeUser(w, http.StatusCreated, u)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if r.Method == http.MethodPut {
		var in struct {
			NamaLengkap string `json:"namaLengkap"`
			Email       string `json:"email"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		if in.NamaLengkap == "" || in.Email == "" {
			writeError(w, http.StatusBadRequest, "namaLengkap dan email wajib diisi")
			return
		}
		updated, err := h.verifier.LocalUpdateProfile(r.Context(), u.UserID, in.NamaLengkap, in.Email)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		u = updated
	}
	h.writeUser(w, http.StatusOK, u)
}

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.verifier.LocalListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	if u != nil && h.audit != nil {
		p := audit.Params{Action: audit.ActionLogout, EntityType: "USER", EntityID: &u.UserID, UserID: u.UserID}
		p.FromRequest(r)
		h.audit.Log(p)
	}
	http.SetCookie(w, &http.Cookie{
		Name: auth.CookieName, Value: "", Path: "/",
		MaxAge: -1, HttpOnly: true, Secure: h.cfg.IsProd,
		SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var in struct{ Password string `json:"password"` }
	if !decodeJSON(w, r, &in) {
		return
	}
	u := auth.FromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if len(in.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Password minimal 6 karakter")
		return
	}
	if err := h.verifier.LocalUpdatePassword(r.Context(), u.UserID, in.Password); err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal mengubah password")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	// Local auth has no email verification; keep the endpoint as a no-op so
	// the FE flow doesn't hard-fail.
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *AuthHandler) writeUser(w http.ResponseWriter, status int, u *auth.AuthUser) {
	writeJSON(w, status, map[string]any{"user": map[string]any{
		"id": u.UserID, "namaLengkap": u.NamaLengkap, "email": u.Email,
		"role": u.Role, "accountType": u.AccountType, "isDemo": u.IsDemo,
	}})
}
