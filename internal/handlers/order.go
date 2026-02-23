package handlers

import (
	"encoding/json"
	"errors"
	"github.com/Guram-Gurych/gophermart.git/internal/middleware"
	"github.com/Guram-Gurych/gophermart.git/internal/repository"
	"github.com/Guram-Gurych/gophermart.git/internal/services"
	"io"
	"log/slog"
	"net/http"
)

type OrderHandler struct {
	service *services.OrderService
	Logger  *slog.Logger
}

func NewOrderHandler(serv *services.OrderService, log *slog.Logger) *OrderHandler {
	return &OrderHandler{service: serv, Logger: log}
}

func (h *OrderHandler) SetOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.WarnContext(r.Context(), "failed to read order request body", slog.Any("error", err))
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := string(b)

	err = h.service.SaveOrder(r.Context(), userID, orderNumber)
	if err != nil {
		if errors.Is(err, services.ErrorInvalidOrderNumber) {
			http.Error(w, "Failed to decode request body", http.StatusUnprocessableEntity)
			return
		} else if errors.Is(err, repository.ErrorOrderAlreadyExists) {
			w.WriteHeader(http.StatusOK)
			return
		} else if errors.Is(err, repository.ErrorOrderConflict) {
			h.Logger.WarnContext(r.Context(), "order conflict: already uploaded by another user",
				slog.String("number", orderNumber),
				slog.String("user_id", userID.String()))
			http.Error(w, "The order number has already been uploaded by another user.", http.StatusConflict)
			return
		} else {
			h.Logger.Error("failed to save order to database",
				slog.String("number", orderNumber),
				slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetOrders(r.Context(), userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to fetch user orders",
			slog.String("user_id", userID.String()),
			slog.Any("error", err))
		http.Error(w, "Error accessing the database", http.StatusInternalServerError)
		return
	} else if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}
