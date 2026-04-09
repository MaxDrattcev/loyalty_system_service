package models

import "time"

type User struct {
	ID             int64     `db:"id"`
	Login          string    `db:"login"`
	Password       string    `db:"password"`
	CurrentBalance int64     `db:"current_balance"`
	TotalWithdrawn int64     `db:"total_withdrawn"`
	CreatedAt      time.Time `db:"created_at"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
