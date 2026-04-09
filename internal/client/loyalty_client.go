package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
)

type loyaltyClient struct {
	cfg    *config.Config
	client *RetryableClient
}

func NewLoyaltyClient(cfg *config.Config) LoyaltyClient {
	return &loyaltyClient{
		cfg:    cfg,
		client: NewRetryableClient(),
	}
}

func (l *loyaltyClient) GetAccrual(ctx context.Context, order models.Order) (models.Accrual, error) {
	url := fmt.Sprintf("http://%s/api/orders/%d", l.cfg.Client.Address, order.Number)

	headers := map[string]string{
		"Content-Length": "0",
	}
	resp, err := l.client.GetWithRetry(ctx, url, headers)
	if err != nil {
		return models.Accrual{}, err
	}
	if resp.StatusCode() == 204 {
		return models.Accrual{}, fmt.Errorf("order %d not found in service loyalty", order.Number)
	}
	accrual := models.Accrual{}
	if err := json.Unmarshal(resp.Body(), &accrual); err != nil {
		return models.Accrual{}, err
	}
	return accrual, nil
}
