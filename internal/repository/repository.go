package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type DBRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *DBRepository {
	return &DBRepository{db: db}
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
