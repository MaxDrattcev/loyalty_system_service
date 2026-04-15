package models

// Accrual represents response from external accrual service.
type Accrual struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}
