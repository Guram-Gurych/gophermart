package orders

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"github.com/google/uuid"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Order, error)
	UpdateOrder(ctx context.Context, orderNumber string, status domain.OrderStatus, accrual domain.JSONBalance) error
	UpdateOrdersStatus(ctx context.Context, numbers []string, status domain.OrderStatus) error
}

type orderService struct {
	repository OrderRepository
}

func NewService(rep OrderRepository) *orderService {
	return &orderService{
		repository: rep,
	}
}

func (s *orderService) SaveOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error {
	if !domain.ValidLuhn(orderNumber) {
		return domain.ErrorInvalidOrderNumber
	}

	return s.repository.SaveOrder(ctx, userID, orderNumber)
}

func (s *orderService) GetOrders(ctx context.Context, userID uuid.UUID) ([]domain.Order, error) {
	return s.repository.GetOrdersByUserID(ctx, userID)
}
