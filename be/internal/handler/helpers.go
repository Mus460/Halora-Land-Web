package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// writeJSON encodes v with the given status. Errors from encoding are ignored.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a uniform {error: msg} body.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

// decodeJSON decodes r.Body into dst. Returns false and writes 400 on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return false
	}
	return true
}

// parseIntParam reads a chi URL param as int32. Writes 400 and returns ok=false
// when missing or non-numeric.
func parseIntParam(w http.ResponseWriter, r *http.Request, name string) (int32, bool) {
	raw := URLParam(r, name)
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		writeError(w, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return int32(v), true
}

// URLParam reads a route parameter from chi's context.
func URLParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

// mapPgErr converts common pgx errors to HTTP statuses.
func mapPgErr(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return http.StatusNotFound, "Not found"
	}
	if strings.Contains(err.Error(), "23505") {
		return http.StatusConflict, "Resource already exists"
	}
	return http.StatusInternalServerError, "Internal server error"
}
