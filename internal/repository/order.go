package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/google/uuid"
)

func (rep *DBRepository) SaveOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error {
	var rowUserID uuid.UUID
	var isInserted bool
	err := rep.db.QueryRowContext(ctx, queryInsertOrder, orderNumber, userID).Scan(&rowUserID, &isInserted)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrorDBReading, err)
	}

	if isInserted {
		return nil
	}

	if rowUserID == userID {
		return ErrorOrderAlreadyExists
	}

	return ErrorOrderConflict
}

func (rep *DBRepository) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return rep.fetchOrders(ctx, "WHERE user_id = $1 ORDER BY uploaded_at DESC", userID)
}

func (rep *DBRepository) UpdateOrder(ctx context.Context, orderNumber string, status models.OrderStatus, accrual models.JSONBalance) error {
	return rep.withTx(ctx, func(tx *sql.Tx) error {
		var userID uuid.UUID
		err := tx.QueryRowContext(ctx, queryUpdateOrderStatusandAccrual, status, accrual, orderNumber).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("%w: %w", ErrorDBWriting, err)
		}

		if status == models.StatusProcessed && accrual > 0 {
			_, err = tx.ExecContext(ctx, queryUpdateBalance, accrual, userID)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrorDBWriting, err)
			}
		}

		return nil
	})
}

func (rep *DBRepository) UpdateOrdersStatus(ctx context.Context, numbers []string, status models.OrderStatus) error {
	_, err := rep.db.ExecContext(ctx, queryUpdateOrderStatus, status, numbers)
	if err != nil {
		return err
	}

	return nil
}

func (rep *DBRepository) fetchOrders(ctx context.Context, condition string, args ...interface{}) ([]models.Order, error) {
	rows, err := rep.db.QueryContext(ctx, querySelectOrdersBase+condition, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorDBReading, err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err = rows.Scan(&order.Number, &order.Status, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrorDataScan, err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorRowsIteration, err)
	}

	return orders, nil
}
