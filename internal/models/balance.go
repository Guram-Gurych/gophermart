package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math"
	"time"
)

type JSONBalance int64

func (b JSONBalance) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.2f", float64(b)/100)), nil
}

func (b *JSONBalance) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}

	*b = JSONBalance(math.Round(f * 100))
	return nil
}

func (b JSONBalance) Value() (driver.Value, error) {
	return int64(b), nil
}

func (b *JSONBalance) Scan(value interface{}) error {
	if value == nil {
		*b = 0
		return nil
	}
	v, ok := value.(int64)
	if !ok {
		return fmt.Errorf("failed to scan balance: %v", value)
	}

	*b = JSONBalance(v)
	return nil
}

type Withdrawal struct {
	UserID      uuid.UUID   `json:"-"`
	OrderNumber string      `json:"order"`
	Sum         JSONBalance `json:"sum"`
	ProcessedAt time.Time   `json:"processed_at,omitempty"`
}

type Balance struct {
	Current   JSONBalance `json:"current"`
	Withdrawn JSONBalance `json:"withdrawn"`
}
