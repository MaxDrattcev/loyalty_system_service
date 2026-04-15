package repository

import (
	"context"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
)

// UserRepository defines persistence operations for users and user balances.
type UserRepository interface {
	// Create inserts a new user and returns created entity with assigned ID.
	Create(ctx context.Context, user models.User) (models.User, error)

	// GetByLogin returns user by login.
	GetByLogin(ctx context.Context, login string) (models.User, error)

	// Update increases user balance using order accrual.
	Update(ctx context.Context, order models.Order) error

	// UpdateTx increases user balance inside provided transaction.
	UpdateTx(ctx context.Context, tx pgx.Tx, order models.Order) error

	// GetByUserID returns user by numeric ID.
	GetByUserID(ctx context.Context, userID int64) (models.User, error)

	// WithdrawIfEnoughTx withdraws amount from balance if funds are sufficient.
	// Returns ErrInsufficientFunds when balance is not enough.
	WithdrawIfEnoughTx(ctx context.Context, tx pgx.Tx, userID int64, amount int64) error
}

// OrderRepository defines persistence operations for orders.
type OrderRepository interface {
	// Create inserts new order record.
	Create(ctx context.Context, order models.Order) error

	// CreateTx inserts new order record within existing transaction.
	CreateTx(ctx context.Context, tx pgx.Tx, order models.Order) error

	// GetByNumber returns order by order number.
	GetByNumber(ctx context.Context, number int64) (models.Order, error)

	// UpdateTx updates order status/accrual within existing transaction.
	UpdateTx(ctx context.Context, tx pgx.Tx, accrual models.Accrual, intAccrual int64) error

	// GetStatusNewAndProcessing returns orders with NEW/PROCESSING status
	// that are old enough for repeated accrual polling.
	GetStatusNewAndProcessing(ctx context.Context) ([]models.Order, error)

	// GetOrdersByUserID returns all orders uploaded by user.
	GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error)
}

// WithdrawalRepository defines persistence operations for withdrawal history.
type WithdrawalRepository interface {
	// CreateTx inserts withdrawal history record within existing transaction.
	CreateTx(ctx context.Context, tx pgx.Tx, wh models.WithdrawalHistory) error

	// GetWithdrawals returns withdrawal history for user.
	GetWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalHistory, error)
}
