package services

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/google/uuid"
)

type OrderService struct {
	repository OrderRepository
}

func NewOrderService(rep OrderRepository) *OrderService {
	return &OrderService{
		repository: rep,
	}
}

func (s *OrderService) SaveOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error {
	if !ValidLuhn(orderNumber) {
		return ErrorInvalidOrderNumber
	}

	return s.repository.SaveOrder(ctx, userID, orderNumber)
}

func (s *OrderService) GetOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return s.repository.GetOrdersByUserID(ctx, userID)
}
