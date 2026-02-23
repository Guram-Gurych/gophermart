package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/google/uuid"
)

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

		withdrawalID, err := uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("%w: %w", ErrorUUIDGenerate, err)
		}

		_, err = tx.ExecContext(ctx, queryInsertWithdrawal, withdrawalID, userID, withdrawal.OrderNumber, withdrawal.Sum)
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
		if err = rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrorDataScan, err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorRowsIteration, err)
	}

	return withdrawals, nil
}
