package models

import "time"

type Order struct {
	ID         int64       `db:"id"`
	Number     int64       `db:"number"`
	UserID     int64       `db:"user_id"`
	Status     OrderStatus `db:"status"`
	Accrual    int64       `db:"accrual"`
	UploadedAt time.Time   `db:"uploaded_at"`
}

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type OrderResponse struct {
	Number   int64       `json:"number"`
	Status   OrderStatus `json:"status"`
	Accrual  *int64      `json:"accrual,omitempty"`
	Uploaded string      `json:"uploaded_at"`
}
