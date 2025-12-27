package repository

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, login, hashPassword string) error
	GetUserByLogin(ctx context.Context, login string) (models.Users, error)
	SaveOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	UpdateOrder(ctx context.Context, orderNumber string, status models.OrderStatus, accrual float64) error
	GetPendingOrders(ctx context.Context) ([]models.Order, error)
	GetBalance(ctx context.Context, userID uuid.UUID) (models.Balance, error)
	SaveWithdraw(ctx context.Context, userID uuid.UUID, withdrawal models.Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error)
}
