package handlers

import (
	"encoding/json"
	"errors"
	"github.com/Guram-Gurych/gophermart.git/internal/repository"
	"github.com/Guram-Gurych/gophermart.git/internal/services"
	"net/http"
)

type userRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(serv *services.AuthService) *AuthHandler {
	return &AuthHandler{service: serv}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user userRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if user.Login == "" || user.Password == "" {
		http.Error(w, "Failed to register the user", http.StatusBadRequest)
		return
	}

	token, err := h.service.Register(r.Context(), user.Login, user.Password)
	if err != nil {
		if errors.Is(err, repository.ErrorUserAlreadyExists) {
			http.Error(w, "Failed to register the user", http.StatusConflict)
			return
		} else {
			http.Error(w, "Failed to register the user", http.StatusInternalServerError)
			return
		}
	}

	h.setAuthCookie(w, token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var user userRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if user.Login == "" || user.Password == "" {
		http.Error(w, "User authorization failed", http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), user.Login, user.Password)
	if err != nil {
		if errors.Is(err, services.ErrorInvalidCredentials) {
			http.Error(w, "User authorization failed", http.StatusUnauthorized)
			return
		} else {
			http.Error(w, "Failed to register the user", http.StatusInternalServerError)
			return
		}
	}

	h.setAuthCookie(w, token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) setAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}
