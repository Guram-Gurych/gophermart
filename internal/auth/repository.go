package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"github.com/Guram-Gurych/gophermart.git/internal/platform/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
)

type repository struct {
	*storage.DBRepository
}

func NewRepository(rep *storage.DBRepository) *repository {
	return &repository{DBRepository: rep}
}

func (rep *repository) CreateUser(ctx context.Context, userID uuid.UUID, login, hashPassword string) error {
	return rep.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, queryInsertUser, userID, login, hashPassword)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return domain.ErrorUserAlreadyExists
				}
			}
			return fmt.Errorf("%w: %w", domain.ErrorDBWriting, err)
		}

		_, err = tx.ExecContext(ctx, queryInsertBalance, userID, 0, 0)
		if err != nil {
			return fmt.Errorf("%w: %w", domain.ErrorDBWriting, err)
		}

		return nil
	})
}

func (rep *repository) GetUserByLogin(ctx context.Context, login string) (domain.Users, error) {
	user := domain.Users{Login: login}
	err := rep.DB.QueryRowContext(ctx, queryGetUserByLogin, login).Scan(&user.ID, &user.HashPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Users{}, domain.ErrorUserNotFound
		}
		return domain.Users{}, fmt.Errorf("%w: %w", domain.ErrorDBReading, err)
	}

	return user, nil
}
