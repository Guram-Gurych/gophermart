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

func (rep *DBRepository) CreateUser(ctx context.Context, userID uuid.UUID, login, hashPassword string) error {
	return rep.withTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, queryInsertUser, userID, login, hashPassword)
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
