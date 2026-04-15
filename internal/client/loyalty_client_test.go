package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/stretchr/testify/require"
)

func newTestLoyaltyClient(address string) *loyaltyClient {
	return &loyaltyClient{
		cfg: &config.Config{
			Client: config.ClientConfig{
				Address: address,
			},
		},
		client: NewRetryableClient(),
	}
}

func TestLoyaltyClient_GetAccrual(t *testing.T) {
	oldRetryDelays := retryDelays
	retryDelays = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	t.Cleanup(func() {
		retryDelays = oldRetryDelays
	})

	order := models.Order{Number: 79927398713}

	t.Run("success: 200 + valid json", func(t *testing.T) {
		var requestedPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestedPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order":"79927398713","status":"PROCESSED","accrual":12.34}`))
		}))
		defer srv.Close()

		c := newTestLoyaltyClient(srv.URL)
		got, err := c.GetAccrual(context.Background(), order)

		require.NoError(t, err)
		require.Equal(t, "/api/orders/79927398713", requestedPath)
		require.Equal(t, "79927398713", got.OrderNumber)
		require.Equal(t, "PROCESSED", got.Status)
		require.Equal(t, 12.34, got.Accrual)
	})

	t.Run("normalizes address without scheme", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order":"79927398713","status":"REGISTERED","accrual":0}`))
		}))
		defer srv.Close()

		trimmed := strings.TrimPrefix(srv.URL, "http://")
		c := newTestLoyaltyClient(trimmed)

		got, err := c.GetAccrual(context.Background(), order)
		require.NoError(t, err)
		require.Equal(t, "REGISTERED", got.Status)
	})

	t.Run("204 after retries -> not found error", func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := newTestLoyaltyClient(srv.URL)
		_, err := c.GetAccrual(context.Background(), order)

		require.Error(t, err)
		require.Contains(t, err.Error(), "not found in service loyalty")
		require.GreaterOrEqual(t, calls.Load(), int32(2))
	})

	t.Run("invalid json -> unmarshal error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order":`))
		}))
		defer srv.Close()

		c := newTestLoyaltyClient(srv.URL)
		_, err := c.GetAccrual(context.Background(), order)

		require.Error(t, err)
	})

	t.Run("transport error (connection refused)", func(t *testing.T) {
		c := newTestLoyaltyClient("127.0.0.1:1")

		_, err := c.GetAccrual(context.Background(), order)
		require.Error(t, err)
	})
}
