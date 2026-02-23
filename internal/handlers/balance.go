package handlers

import (
	"encoding/json"
	"errors"
	"github.com/Guram-Gurych/gophermart.git/internal/middleware"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/Guram-Gurych/gophermart.git/internal/repository"
	"github.com/Guram-Gurych/gophermart.git/internal/services"
	"log/slog"
	"net/http"
)

type BalanceHandler struct {
	service *services.BalanceService
	Logger  *slog.Logger
}

func NewBalanceHandler(serv *services.BalanceService, log *slog.Logger) *BalanceHandler {
	return &BalanceHandler{service: serv, Logger: log}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.service.GetBalance(r.Context(), userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get user balance",
			slog.String("user_id", userID.String()),
			slog.Any("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(balance); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to encode balance response", slog.Any("error", err))
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

func (h *BalanceHandler) SetWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	var withdrawal models.Withdrawal
	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err := h.service.Withdraw(r.Context(), userID, withdrawal)
	if err != nil {
		if errors.Is(err, services.ErrorInvalidOrderNumber) {
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		} else if errors.Is(err, repository.ErrorInsufficientFunds) {
			http.Error(w, "There are insufficient funds in the account", http.StatusPaymentRequired)
			return
		} else {
			h.Logger.ErrorContext(r.Context(), "withdrawal service failure",
				slog.String("user_id", userID.String()),
				slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetWithdraws(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	withdraws, err := h.service.GetWithdrawals(r.Context(), userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to fetch withdrawals",
			slog.String("user_id", userID.String()),
			slog.Any("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdraws) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(withdraws); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to encode withdrawals list", slog.Any("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
