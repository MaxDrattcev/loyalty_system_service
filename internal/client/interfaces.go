package client

import (
	"context"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
)

// LoyaltyClient fetches accrual information for orders from external accrual service.
type LoyaltyClient interface {
	// GetAccrual returns accrual data for the specified order.
	// It returns an error when external service request fails or response cannot be parsed.
	GetAccrual(ctx context.Context, order models.Order) (models.Accrual, error)
}
