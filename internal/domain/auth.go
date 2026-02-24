package domain

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Users struct {
	ID           uuid.UUID `json:"-"`
	Login        string    `json:"login"`
	HashPassword string    `json:"password"`
}

type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}
