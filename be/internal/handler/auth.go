package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/halora-land/halora-be/internal/audit"
	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/config"
)

// AuthHandler proxies Supabase auth and issues a consistent cookie
// (ARCHITECTURE.md §3.3, §7.2). Logout reliably clears via MaxAge=-1.
type AuthHandler struct {
	cfg      *config.Config
	verifier *auth.Verifier
	audit    *audit.Logger
	http     *http.Client
}

func NewAuthHandler(cfg *config.Config, v *auth.Verifier, al *audit.Logger) *AuthHandler {
	return &AuthHandler{cfg: cfg, verifier: v, audit: al, http: &http.Client{Timeout: 15 * time.Second}}
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

	// Demo mode: local DB auth — verify email/password, issue session JWT.
	if h.verifier.DemoMode() {
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
			Name: h.cfg.CookieName, Value: token, Path: "/",
			MaxAge: 86400, HttpOnly: true, Secure: h.cfg.IsProd,
			SameSite: http.SameSiteStrictMode,
		})
		if h.audit != nil {
			p := audit.Params{Action: audit.ActionLogin, EntityType: "USER", EntityID: &u.UserID, UserID: u.UserID}
			p.FromRequest(r)
			h.audit.Log(p)
		}
		writeJSON(w, http.StatusOK, map[string]any{"user": map[string]any{
			"id": u.UserID, "namaLengkap": u.NamaLengkap, "email": u.Email,
			"role": u.Role, "accountType": u.AccountType, "isDemo": u.IsDemo,
		}})
		return
	}

	if in.Email == "" || in.Password == "" {
		writeError(w, http.StatusBadRequest, "email dan password wajib diisi")
		return
	}
	body, _ := json.Marshal(map[string]any{
		"email": in.Email, "password": in.Password,
	})
	resp, err := h.supabase("/auth/v1/token?grant_type=password", body, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Email atau password salah")
		return
	}
	defer resp.Body.Close()
	sb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		writeJSON(w, resp.StatusCode, json.RawMessage(sb))
		return
	}
	var sess struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		ExpiresAt    int64  `json:"expires_at"`
		TokenType    string `json:"token_type"`
		User         struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(sb, &sess); err != nil {
		writeError(w, http.StatusInternalServerError, "Invalid session response")
		return
	}
	cookieVal, _ := json.Marshal(map[string]any{
		"access_token":  sess.AccessToken,
		"refresh_token": sess.RefreshToken,
		"expires_at":    sess.ExpiresAt,
		"expires_in":    sess.ExpiresIn,
		"token_type":    sess.TokenType,
		"user":          sess.User,
	})
	http.SetCookie(w, &http.Cookie{
		Name: h.cfg.CookieName, Value: string(cookieVal), Path: "/",
		MaxAge: sess.ExpiresIn, HttpOnly: true, Secure: h.cfg.IsProd,
		SameSite: http.SameSiteStrictMode,
	})

	u, err := h.verifier.LoadOrCreate(r.Context(), sess.User.ID, sess.User.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Gagal memuat user")
		return
	}
	if h.audit != nil {
		p := audit.Params{Action: audit.ActionLogin, EntityType: "USER", EntityID: &u.UserID, UserID: u.UserID}
		p.FromRequest(r)
		h.audit.Log(p)
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": map[string]any{
		"id": u.UserID, "namaLengkap": u.NamaLengkap, "email": u.Email,
		"role": u.Role, "accountType": u.AccountType, "isDemo": u.IsDemo,
	}})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		NamaLengkap string `json:"namaLengkap"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}

	// Demo mode: insert directly into local DB with hashed password.
	if h.verifier.DemoMode() {
		if in.Email == "" || in.Password == "" || in.NamaLengkap == "" {
			writeError(w, http.StatusBadRequest, "namaLengkap, email, dan password wajib diisi")
			return
		}
		if len(in.Password) < 6 {
			writeError(w, http.StatusBadRequest, "Password minimal 6 karakter")
			return
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
		writeJSON(w, http.StatusCreated, map[string]any{"user": map[string]any{
			"id": u.UserID, "namaLengkap": u.NamaLengkap, "email": u.Email,
			"role": u.Role, "accountType": u.AccountType, "isDemo": u.IsDemo,
		}})
		return
	}

	body, _ := json.Marshal(map[string]any{
		"email": in.Email, "password": in.Password,
		"data":  map[string]any{"namaLengkap": in.NamaLengkap},
	})
	resp, err := h.supabase("/auth/v1/signup", body, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer resp.Body.Close()
	sb, _ := io.ReadAll(resp.Body)
	if au := auth.FromContext(r.Context()); h.audit != nil && au != nil {
		p := audit.Params{Action: audit.ActionRegister, EntityType: "USER", EntityID: &au.UserID, UserID: au.UserID}
		p.FromRequest(r)
		h.audit.Log(p)
	}
	writeJSON(w, resp.StatusCode, json.RawMessage(sb))
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
		if h.verifier.DemoMode() {
			updated, err := h.verifier.LocalUpdateProfile(r.Context(), u.UserID, in.NamaLengkap, in.Email)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			u = updated
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": map[string]any{
		"id": u.UserID, "namaLengkap": u.NamaLengkap, "email": u.Email,
		"role": u.Role, "accountType": u.AccountType, "isDemo": u.IsDemo,
	}})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	if u != nil {
		if !h.verifier.DemoMode() {
			if token := bearer(r); token != "" {
				body, _ := json.Marshal(map[string]any{})
				req, _ := http.NewRequest(http.MethodPost, h.cfg.SupabaseURL+"/auth/v1/logout", bytes.NewReader(body))
				req.Header.Set("apikey", h.cfg.SupabaseAnonKey)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")
				_, _ = h.http.Do(req)
			}
		}
		if h.audit != nil {
			p := audit.Params{Action: audit.ActionLogout, EntityType: "USER", EntityID: &u.UserID, UserID: u.UserID}
			p.FromRequest(r)
			h.audit.Log(p)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: h.cfg.CookieName, Value: "", Path: "/",
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

	// Demo mode: update password hash directly in local DB.
	if h.verifier.DemoMode() {
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
		return
	}

	token := bearer(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	body, _ := json.Marshal(map[string]any{"password": in.Password})
	resp, err := h.supabaseAuthed("/auth/v1/user", body, "PUT", token)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer resp.Body.Close()
	sb, _ := io.ReadAll(resp.Body)
	writeJSON(w, resp.StatusCode, json.RawMessage(sb))
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email string `json:"email"` }
	if !decodeJSON(w, r, &in) {
		return
	}

	// Demo mode: no email verification — just return success.
	if h.verifier.DemoMode() {
		writeJSON(w, http.StatusOK, map[string]any{"success": true})
		return
	}

	body, _ := json.Marshal(map[string]any{"type": "signup", "email": in.Email})
	resp, err := h.supabase("/auth/v1/resend", body, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer resp.Body.Close()
	sb, _ := io.ReadAll(resp.Body)
	writeJSON(w, resp.StatusCode, json.RawMessage(sb))
}

func (h *AuthHandler) supabase(path string, body []byte, r *http.Request) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodPost, h.cfg.SupabaseURL+path, bytes.NewReader(body))
	req.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	req.Header.Set("Content-Type", "application/json")
	return h.http.Do(req)
}

func (h *AuthHandler) supabaseAuthed(path string, body []byte, method, token string) (*http.Response, error) {
	req, _ := http.NewRequest(method, h.cfg.SupabaseURL+path, bytes.NewReader(body))
	req.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return h.http.Do(req)
}

func bearer(r *http.Request) string {
	if h := r.Header.Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
		return h[7:]
	}
	return ""
}
