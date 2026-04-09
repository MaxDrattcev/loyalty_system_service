package service

import (
	"context"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
)

type UserService interface {
	Register(ctx context.Context, user models.User) (string, error)

	Login(ctx context.Context, user models.User) (string, error)

	GetBalance(ctx context.Context, userID int64) (models.BalanceResponse, error)
}

type OrderService interface {
	Create(ctx context.Context, numberOrder, userID int64) error

	GetOrders(ctx context.Context, userID int64) ([]models.OrderResponse, error)
}

type WithdrawalService interface {
	Withdrawal(ctx context.Context, userID, order int64, sum float64) error

	GetWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalResponse, error)
}
