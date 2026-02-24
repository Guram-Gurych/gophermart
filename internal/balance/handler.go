package balance

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (domain.Balance, error)
	Withdraw(ctx context.Context, userID uuid.UUID, w domain.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]domain.Withdrawal, error)
}

type balanceHandler struct {
	service BalanceService
	Logger  *slog.Logger
}

func NewBalanceHandler(serv BalanceService, log *slog.Logger) *balanceHandler {
	return &balanceHandler{service: serv, Logger: log}
}

func (h *balanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := domain.GetUserID(r.Context())
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

func (h *balanceHandler) SetWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := domain.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	var withdrawal domain.Withdrawal
	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err := h.service.Withdraw(r.Context(), userID, withdrawal)
	if err != nil {
		if errors.Is(err, domain.ErrorInvalidOrderNumber) {
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		} else if errors.Is(err, domain.ErrorInsufficientFunds) {
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

func (h *balanceHandler) GetWithdraws(w http.ResponseWriter, r *http.Request) {
	userID, ok := domain.GetUserID(r.Context())
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
