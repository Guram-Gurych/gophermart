package domain

import (
	"context"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	RequestIDKey contextKey = "requestIDKey"
)

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}
