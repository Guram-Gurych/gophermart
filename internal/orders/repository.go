package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"github.com/Guram-Gurych/gophermart.git/internal/platform/storage"
	"github.com/google/uuid"
)

type repository struct {
	*storage.DBRepository
}

func NewRepository(rep *storage.DBRepository) *repository {
	return &repository{DBRepository: rep}
}

func (rep *repository) SaveOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error {
	var rowUserID uuid.UUID
	var isInserted bool
	err := rep.DB.QueryRowContext(ctx, queryInsertOrder, orderNumber, userID).Scan(&rowUserID, &isInserted)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrorDBReading, err)
	}

	if isInserted {
		return nil
	}

	if rowUserID == userID {
		return domain.ErrorOrderAlreadyExists
	}

	return domain.ErrorOrderConflict
}

func (rep *repository) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Order, error) {
	return rep.fetchOrders(ctx, "WHERE user_id = $1 ORDER BY uploaded_at DESC", userID)
}

func (rep *repository) UpdateOrder(ctx context.Context, orderNumber string, status domain.OrderStatus, accrual domain.JSONBalance) error {
	return rep.WithTx(ctx, func(tx *sql.Tx) error {
		var userID uuid.UUID
		err := tx.QueryRowContext(ctx, queryUpdateOrderStatusandAccrual, status, accrual, orderNumber).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("%w: %w", domain.ErrorDBWriting, err)
		}

		if status == domain.StatusProcessed && accrual > 0 {
			_, err = tx.ExecContext(ctx, queryUpdateBalance, accrual, userID)
			if err != nil {
				return fmt.Errorf("%w: %w", domain.ErrorDBWriting, err)
			}
		}

		return nil
	})
}

func (rep *repository) UpdateOrdersStatus(ctx context.Context, numbers []string, status domain.OrderStatus) error {
	_, err := rep.DB.ExecContext(ctx, queryUpdateOrderStatus, status, numbers)
	if err != nil {
		return err
	}

	return nil
}

func (rep *repository) GetPendingOrders(ctx context.Context, limit int) ([]string, error) {
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

func (rep *repository) fetchOrders(ctx context.Context, condition string, args ...interface{}) ([]domain.Order, error) {
	rows, err := rep.DB.QueryContext(ctx, querySelectOrdersBase+condition, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrorDBReading, err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		if err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrorDataScan, err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrorRowsIteration, err)
	}

	return orders, nil
}
