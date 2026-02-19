package models

import "github.com/google/uuid"

type Users struct {
	ID           uuid.UUID `json:"-"`
	Login        string    `json:"login"`
	HashPassword string    `json:"password"`
}
