package domain

import (
	"database/sql/driver"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))
	return []byte(stamp), nil
}

func (t *JSONTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if s == "" || s == "null" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = JSONTime(parsed)
	return nil
}

func (t JSONTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

func (t *JSONTime) Scan(value interface{}) error {
	if value == nil {
		*t = JSONTime(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*t = JSONTime(v)
	case []byte:
		parsed, err := time.Parse("2006-01-02 15:04:05.999999-07", string(v))
		if err != nil {
			return err
		}
		*t = JSONTime(parsed)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	return nil
}

type OrderStatus string

const (
	StatusREGISTERED OrderStatus = "REGISTERED"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     string       `json:"number"`
	UserID     uuid.UUID    `json:"-"`
	Status     OrderStatus  `json:"status"`
	Accrual    *JSONBalance `json:"accrual,omitempty"`
	UploadedAt JSONTime     `json:"uploaded_at"`
}
