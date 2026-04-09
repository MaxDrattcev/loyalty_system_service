package repository

import (
	"context"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) (models.User, error)

	GetByLogin(ctx context.Context, login string) (models.User, error)

	Update(ctx context.Context, order models.Order) error

	UpdateTx(ctx context.Context, tx pgx.Tx, order models.Order) error

	GetByUserID(ctx context.Context, userID int64) (models.User, error)

	WithdrawIfEnoughTx(ctx context.Context, tx pgx.Tx, userID int64, amount int64) error
}

type OrderRepository interface {
	Create(ctx context.Context, order models.Order) error

	CreateTx(ctx context.Context, tx pgx.Tx, order models.Order) error

	GetByNumber(ctx context.Context, number int64) (models.Order, error)

	Update(ctx context.Context, accrual models.Accrual, intAccrual int64) error

	UpdateTx(ctx context.Context, tx pgx.Tx, accrual models.Accrual, intAccrual int64) error

	GetStatusNewAndProcessing(ctx context.Context) ([]models.Order, error)

	GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error)
}

type WithdrawalRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, wh models.WithdrawalHistory) error

	GetWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalHistory, error)
}
