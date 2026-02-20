package repository

import (
	"context"
)

func (rep *DBRepository) GetPendingOrders(ctx context.Context, limit int) ([]string, error) {
	listOrder, err := rep.fetchOrders(
		ctx,
		"WHERE status in ('NEW', 'REGISTERED', 'PROCESSING') ORDER BY uploaded_at ASC LIMIT $1",
		limit)
	if err != nil {
		return nil, err
	}

	numbersOrders := make([]string, 0, len(listOrder))
	for _, order := range listOrder {
		numbersOrders = append(numbersOrders, order.Number)
	}

	return numbersOrders, nil
}
