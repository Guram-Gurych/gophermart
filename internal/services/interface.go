package services

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/google/uuid"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, userID uuid.UUID, login, hashPassword string) error
	GetUserByLogin(ctx context.Context, login string) (models.Users, error)
}

type OrderRepository interface {
	SaveOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (models.Balance, error)
	SaveWithdraw(ctx context.Context, userID uuid.UUID, withdrawal models.Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error)
}

type WorkerRepository interface {
	GetPendingOrders(ctx context.Context, limit int) ([]string, error)
	UpdateOrder(ctx context.Context, orderNumber string, status models.OrderStatus, accrual models.JSONBalance) error
	UpdateOrdersStatus(ctx context.Context, numbers []string, status models.OrderStatus) error
}
