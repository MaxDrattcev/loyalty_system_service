package service

import (
	"context"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
)

// UserService defines business operations for user registration, authentication, and balance retrieval.
type UserService interface {
	// Register creates a new user and returns signed JWT token.
	// Returns an error if password hashing, persistence, or token creation fails.
	Register(ctx context.Context, user models.User) (string, error)

	// Login authenticates user credentials and returns signed JWT token.
	// Returns an error if user is not found, password is invalid, or token creation fails.
	Login(ctx context.Context, user models.User) (string, error)

	// GetBalance returns current balance and total withdrawn amount for the user.
	// Monetary values are returned in major units (e.g. rubles), not cents.
	GetBalance(ctx context.Context, userID int64) (models.BalanceResponse, error)
}

// OrderService defines business operations for order upload and retrieval.
type OrderService interface {
	// Create stores a new order and schedules it for accrual processing.
	// Returns ErrOrderAlreadyUploadedBySameUser when the same user uploads duplicate order.
	Create(ctx context.Context, numberOrder, userID int64) error

	// GetOrders returns user orders mapped to API response format.
	// Returns ErrNoOrders when user has no orders.
	GetOrders(ctx context.Context, userID int64) ([]models.OrderResponse, error)
}

// WithdrawalService defines business operations for withdrawals and withdrawal history.
type WithdrawalService interface {
	// Withdrawal creates withdrawal operation for user and updates balances atomically.
	// Returns an error when order creation, balance update, or history insert fails.
	Withdrawal(ctx context.Context, userID, order int64, sum float64) error

	// GetWithdrawals returns user withdrawal history mapped to API response format.
	// Returns ErrNoWithdraws when user has no withdrawals.
	GetWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalResponse, error)
}
