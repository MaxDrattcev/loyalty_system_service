package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"strings"
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
	addr := strings.TrimSpace(l.cfg.Client.Address)
	addr = strings.TrimRight(addr, "/")

	if strings.HasPrefix(addr, "http//") {
		addr = "http://" + strings.TrimPrefix(addr, "http//")
	}

	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	url := fmt.Sprintf("%s/api/orders/%d", addr, order.Number)

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
