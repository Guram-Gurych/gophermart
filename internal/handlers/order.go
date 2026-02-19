package handlers

import (
	"encoding/json"
	"errors"
	"github.com/Guram-Gurych/gophermart.git/internal/middleware"
	"github.com/Guram-Gurych/gophermart.git/internal/repository"
	"github.com/Guram-Gurych/gophermart.git/internal/services"
	"io"
	"net/http"
)

type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler(serv *services.OrderService) *OrderHandler {
	return &OrderHandler{service: serv}
}

func (h *OrderHandler) SetOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = h.service.SaveOrder(r.Context(), userID, string(b))
	if err != nil {
		if errors.Is(err, services.ErrorInvalidOrderNumber) {
			http.Error(w, "Failed to decode request body", http.StatusUnprocessableEntity)
			return
		} else if errors.Is(err, repository.ErrorOrderAlreadyExists) {
			w.WriteHeader(http.StatusOK)
			return
		} else if errors.Is(err, repository.ErrorOrderConflict) {
			http.Error(w, "The order number has already been uploaded by another user.", http.StatusConflict)
			return
		} else {
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
