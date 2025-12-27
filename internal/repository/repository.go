package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
)

type DBRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *DBRepository {
	return &DBRepository{db: db}
}

func (rep *DBRepository) CreateUser(ctx context.Context, login, hashPassword string) error {
	return rep.withTx(ctx, func(tx *sql.Tx) error {
		userID, err := uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("%w: %w", ErrorUUIDGenerate, err)
		}

		_, err = tx.ExecContext(ctx, queryInsertUser, userID, login, hashPassword)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return ErrorUserAlreadyExists
				}
			}
			return fmt.Errorf("%w: %w", ErrorDBWriting, err)
		}

		_, err = tx.ExecContext(ctx, queryInsertBalance, userID, 0, 0)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrorDBWriting, err)
		}

		return nil
	})
}

func (rep *DBRepository) GetUserByLogin(ctx context.Context, login string) (models.Users, error) {
	user := models.Users{Login: login}
	err := rep.db.QueryRowContext(ctx, queryGetUserByLogin, login).Scan(&user.ID, &user.HashPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Users{}, ErrorUserNotFound
		}
		return models.Users{}, fmt.Errorf("%w: %w", ErrorDBReading, err)
	}

	return user, nil
}

func (rep *DBRepository) SaveOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error {
	_, err := rep.db.ExecContext(ctx, queryInsertOrder, orderNumber, userID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			var existsUserID uuid.UUID
			rowErr := rep.db.QueryRowContext(ctx, querySelectOrder, orderNumber).Scan(&existsUserID)
			if rowErr != nil {
				return fmt.Errorf("%w: %w", ErrorDBReading, err)
			}

			if existsUserID == userID {
				return ErrorOrderAlreadyExists
			}
			return ErrorOrderConflict
		} else {
			return fmt.Errorf("%w: %w", err)
		}
	}

	return nil
}

func (rep *DBRepository) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return rep.fetchOrders(ctx, "user_id = $1", userID)
}

func (rep *DBRepository) UpdateOrder(ctx context.Context, orderNumber string, status models.OrderStatus, accrual float64) error {
	return rep.withTx(ctx, func(tx *sql.Tx) error {
		var userID uuid.UUID
		err := tx.QueryRowContext(ctx, queryUpdateOrderStatus, status, accrual, orderNumber).Scan(&userID)
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

func (rep *DBRepository) GetPendingOrders(ctx context.Context) ([]models.Order, error) {
	return rep.fetchOrders(ctx, "status IN ('NEW', 'PROCESSING')")
}

func (rep *DBRepository) GetBalance(ctx context.Context, userID uuid.UUID) (models.Balance, error) {
	var balance models.Balance
	err := rep.db.QueryRowContext(ctx, querySelectBalance, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Balance{Current: 0, Withdrawn: 0}, nil
		}
		return balance, fmt.Errorf("%w: %w", ErrorDBReading, err)
	}

	return balance, nil
}

func (rep *DBRepository) SaveWithdraw(ctx context.Context, userID uuid.UUID, withdrawal models.Withdrawal) error {
	return rep.withTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, queryUpdateBalanceDecrease, withdrawal.Sum, userID)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrorDBWriting, err)
		}

		if count, err := res.RowsAffected(); count == 0 && err == nil {
			return ErrorInsufficientFunds
		} else if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, queryInsertWithdrawal, userID, withdrawal.OrderNumber, withdrawal.Sum)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrorDBWriting, err)
		}

		return nil
	})

}

func (rep *DBRepository) GetWithdrawalsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error) {
	rows, err := rep.db.QueryContext(ctx, querySelectWithdrawalsByUser, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorDBReading, err)
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var withdrawal models.Withdrawal
		if err := rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrorDataScan, err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorRowsIteration, err)
	}

	return withdrawals, nil
}

func (rep *DBRepository) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := rep.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrorTransactionStart, err)
	}

	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
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
		if err := rows.Scan(&order.Number, &order.Status, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrorDataScan, err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorRowsIteration, err)
	}

	return orders, nil
}
