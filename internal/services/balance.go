package services

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/google/uuid"
)

type BalanceService struct {
	repository BalanceRepository
}

func NewBalanceService(rep BalanceRepository) *BalanceService {
	return &BalanceService{repository: rep}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID uuid.UUID) (models.Balance, error) {
	return s.repository.GetBalance(ctx, userID)
}

func (s *BalanceService) Withdraw(ctx context.Context, userID uuid.UUID, w models.Withdrawal) error {
	if !ValidLuhn(w.OrderNumber) {
		return ErrorInvalidOrderNumber
	}

	return s.repository.SaveWithdraw(ctx, userID, w)
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error) {
	return s.repository.GetWithdrawalsByUserID(ctx, userID)
}
