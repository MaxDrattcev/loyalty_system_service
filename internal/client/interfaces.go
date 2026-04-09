package client

import (
	"context"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
)

type LoyaltyClient interface {
	GetAccrual(ctx context.Context, order models.Order) (models.Accrual, error)
}
