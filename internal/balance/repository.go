package balance

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

func (rep *repository) GetBalance(ctx context.Context, userID uuid.UUID) (domain.Balance, error) {
	var balance domain.Balance
	err := rep.DB.QueryRowContext(ctx, querySelectBalance, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Balance{Current: 0, Withdrawn: 0}, nil
		}
		return balance, fmt.Errorf("%w: %w", domain.ErrorDBReading, err)
	}

	return balance, nil
}

func (rep *repository) SaveWithdraw(ctx context.Context, userID uuid.UUID, withdrawal domain.Withdrawal) error {
	return rep.WithTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, queryUpdateBalanceDecrease, withdrawal.Sum, userID)
		if err != nil {
			return fmt.Errorf("%w: %w", domain.ErrorDBWriting, err)
		}

		if count, err := res.RowsAffected(); count == 0 && err == nil {
			return domain.ErrorInsufficientFunds
		} else if err != nil {
			return err
		}

		withdrawalID, err := uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("%w: %w", domain.ErrorUUIDGenerate, err)
		}

		_, err = tx.ExecContext(ctx, queryInsertWithdrawal, withdrawalID, userID, withdrawal.OrderNumber, withdrawal.Sum)
		if err != nil {
			return fmt.Errorf("%w: %w", domain.ErrorDBWriting, err)
		}

		return nil
	})
}

func (rep *repository) GetWithdrawalsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Withdrawal, error) {
	rows, err := rep.DB.QueryContext(ctx, querySelectWithdrawalsByUser, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrorDBReading, err)
	}
	defer rows.Close()

	var withdrawals []domain.Withdrawal
	for rows.Next() {
		var withdrawal domain.Withdrawal
		if err = rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrorDataScan, err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrorRowsIteration, err)
	}

	return withdrawals, nil
}
