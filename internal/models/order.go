package models

import "time"

type Order struct {
	ID         int       `json:"-"`
	Number     string    `json:"number"`
	UserID     int       `json:"-"`
	Status     string    `json:"status"` // статус обработки (NEW, PROCESSING, INVALID, PROCESSED)
	UploadedAt time.Time `json:"uploaded_at"`
	Accrual    float64   `json:"accrual,omitempty"`
}
