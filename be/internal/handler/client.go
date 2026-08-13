package handler

import (
	"net/http"

	"github.com/halora-land/halora-be/internal/auth"
	"github.com/halora-land/halora-be/internal/repository"
)

type ClientHandler struct {
	repo *repository.ClientRepo
}

func NewClientHandler(repo *repository.ClientRepo) *ClientHandler {
	return &ClientHandler{repo: repo}
}

func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	out, err := h.repo.List(r.Context(), u.UserID, r.URL.Query().Get("search"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"clients": out})
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := auth.FromContext(r.Context())
	var in struct {
		Name    string  `json:"name"`
		Address *string `json:"address"`
		Contact *string `json:"contact"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" {
		writeError(w, http.StatusBadRequest, "name wajib diisi")
		return
	}
	m, err := h.repo.Create(r.Context(), repository.CreateClientInput{
		Name: in.Name, Address: in.Address, Contact: in.Contact, UserID: u.UserID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	u := auth.FromContext(r.Context())
	var in struct {
		Name    *string `json:"name"`
		Address *string `json:"address"`
		Contact *string `json:"contact"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Name != nil && *in.Name == "" {
		writeError(w, http.StatusBadRequest, "name wajib diisi")
		return
	}
	m, err := h.repo.Update(r.Context(), id, u.UserID, repository.UpdateClientInput{
		Name: in.Name, Address: in.Address, Contact: in.Contact,
	})
	if err != nil {
		st, msg := mapPgErr(err)
		writeError(w, st, msg)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntParam(w, r, "id")
	if !ok {
		return
	}
	u := auth.FromContext(r.Context())
	found, err := h.repo.Delete(r.Context(), id, u.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "Klien tidak ditemukan")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}