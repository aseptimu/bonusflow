package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aseptimu/internal/services"
	"github.com/aseptimu/internal/utils"
)

type UserHandler struct {
	authService services.UserService
}

func NewUserHandler(authService services.UserService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

type AuthUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AuthUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogWithError(r.Context(), "RegisterUser: error decoding json", err)
		http.Error(w, "Error decoding json: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.authService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrorUserAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AuthUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogWithError(r.Context(), "LoginUser: error decoding json", err)
		http.Error(w, "Error decoding json: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
