package models

import "time"

// Order represents order entity stored in database.
type Order struct {
	ID         int64       `db:"id"`
	Number     int64       `db:"number"`
	UserID     int64       `db:"user_id"`
	Status     OrderStatus `db:"status"`
	Accrual    int64       `db:"accrual"`
	UploadedAt time.Time   `db:"uploaded_at"`
}

// OrderStatus represents processing state of an order.
type OrderStatus string

const (
	// OrderStatusNew indicates a newly uploaded order.
	OrderStatusNew OrderStatus = "NEW"
	// OrderStatusProcessing indicates an order currently being processed by accrual system.
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusInvalid indicates an invalid order number.
	OrderStatusInvalid OrderStatus = "INVALID"
	// OrderStatusProcessed indicates a fully processed order with final accrual.
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

// OrderResponse represents API response payload for user orders.
type OrderResponse struct {
	Number   string      `json:"number"`
	Status   OrderStatus `json:"status"`
	Accrual  *float64    `json:"accrual,omitempty"`
	Uploaded string      `json:"uploaded_at"`
}
