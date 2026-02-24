package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
)

type DBRepository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *DBRepository {
	return &DBRepository{DB: db}
}

func (rep *DBRepository) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := rep.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrorTransactionStart, err)
	}

	defer tx.Rollback()

	if err = fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
