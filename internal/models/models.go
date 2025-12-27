package models

import (
	"github.com/google/uuid"
	"time"
)

type OrderStatus string

const (
	StatusNew        OrderStatus = "NEW"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessed  OrderStatus = "PROCESSED"
)

type Users struct {
	ID           uuid.UUID
	Login        string
	HashPassword string
}

type Order struct {
	Number     string
	UserID     uuid.UUID
	Status     OrderStatus
	Accrual    float64
	UploadedAt time.Time
}

type Withdrawal struct {
	UserID      uuid.UUID
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

type Balance struct {
	Current   float64
	Withdrawn float64
}
