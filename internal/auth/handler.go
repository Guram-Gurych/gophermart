package auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"log/slog"
	"net/http"
)

type AuthService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

type userRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authHandler struct {
	service AuthService
	Logger  *slog.Logger
}

func NewAuthHandler(serv AuthService, log *slog.Logger) *authHandler {
	return &authHandler{service: serv, Logger: log}
}

func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, domain.ErrorUserAlreadyExists) {
			h.Logger.InfoContext(r.Context(), "registration failed: user already exists", slog.String("login", user.Login))
			http.Error(w, "Failed to register the user", http.StatusConflict)
			return
		} else {
			h.Logger.ErrorContext(r.Context(), "internal error during registration",
				slog.String("login", user.Login),
				slog.Any("error", err))
			http.Error(w, "Failed to register the user", http.StatusInternalServerError)
			return
		}
	}

	h.setAuthCookie(w, token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, domain.ErrorInvalidCredentials) {
			h.Logger.WarnContext(r.Context(), "auth failed: invalid credentials", slog.String("login", user.Login))
			http.Error(w, "User authorization failed", http.StatusUnauthorized)
			return
		} else {
			h.Logger.ErrorContext(r.Context(), "login service internal error",
				slog.String("login", user.Login),
				slog.Any("error", err))
			http.Error(w, "Failed to register the user", http.StatusInternalServerError)
			return
		}
	}

	h.setAuthCookie(w, token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *authHandler) setAuthCookie(w http.ResponseWriter, token string) {
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
