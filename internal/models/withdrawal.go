package models

import "time"

type WithdrawalHistory struct {
	ID            int64     `db:"id"`
	UserID        int64     `db:"user_id"`
	OrderNumber   int64     `db:"order_number"`
	SumWithdrawal int64     `db:"sum_withdrawal"`
	ProcessedAt   time.Time `db:"processed_at"`
}

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt *string `json:"processed_at,omitempty"`
}
