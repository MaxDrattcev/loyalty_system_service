package models

import "time"

type User struct {
	ID             int64     `db:"id" json:"-"`
	Login          string    `db:"login" json:"login"`
	Password       string    `db:"password" json:"password"`
	CurrentBalance int64     `db:"current_balance" json:"-"`
	TotalWithdrawn int64     `db:"total_withdrawn" json:"-"`
	CreatedAt      time.Time `db:"created_at" json:"-"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
