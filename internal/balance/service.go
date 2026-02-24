package balance

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"github.com/google/uuid"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (domain.Balance, error)
	SaveWithdraw(ctx context.Context, userID uuid.UUID, withdrawal domain.Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Withdrawal, error)
}

type service struct {
	repository BalanceRepository
}

func NewService(rep BalanceRepository) *service {
	return &service{repository: rep}
}

func (s *service) GetBalance(ctx context.Context, userID uuid.UUID) (domain.Balance, error) {
	return s.repository.GetBalance(ctx, userID)
}

func (s *service) Withdraw(ctx context.Context, userID uuid.UUID, w domain.Withdrawal) error {
	if !domain.ValidLuhn(w.OrderNumber) {
		return domain.ErrorInvalidOrderNumber
	}

	return s.repository.SaveWithdraw(ctx, userID, w)
}

func (s *service) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]domain.Withdrawal, error) {
	return s.repository.GetWithdrawalsByUserID(ctx, userID)
}
