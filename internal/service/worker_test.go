package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/mocks"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWorker_BuildSnapshot(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		userRepo := mocks.NewMock_UserRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)

		expected := []models.Order{
			{ID: 1, Number: 79927398713, UserID: 10, Status: models.OrderStatusNew},
			{ID: 2, Number: 4242424242424242, UserID: 11, Status: models.OrderStatusProcessing},
		}

		orderRepo.EXPECT().
			GetStatusNewAndProcessing(mock.Anything).
			Return(expected, nil).
			Once()

		w := &Worker{
			cfg:           &config.Config{},
			loyaltyClient: loyaltyClient,
			orderRepo:     orderRepo,
			userRepo:      userRepo,
			jobs:          make(chan models.Order, 10),
		}

		got := w.BuildSnapshot(context.Background())
		require.Equal(t, expected, got)
	})

	t.Run("repo error returns nil", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		userRepo := mocks.NewMock_UserRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)

		orderRepo.EXPECT().
			GetStatusNewAndProcessing(mock.Anything).
			Return(nil, errors.New("db error")).
			Once()

		w := &Worker{
			cfg:           &config.Config{},
			loyaltyClient: loyaltyClient,
			orderRepo:     orderRepo,
			userRepo:      userRepo,
			jobs:          make(chan models.Order, 10),
		}

		got := w.BuildSnapshot(context.Background())
		require.Nil(t, got)
	})
}

func TestWorker_ProcessAccrual(t *testing.T) {
	order := models.Order{
		ID:     1,
		Number: 79927398713,
		UserID: 10,
		Status: models.OrderStatusNew,
	}

	t.Run("loyalty client error", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		userRepo := mocks.NewMock_UserRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)

		loyaltyClient.EXPECT().
			GetAccrual(mock.Anything, order).
			Return(models.Accrual{}, errors.New("client error")).
			Once()

		w := &Worker{
			loyaltyClient: loyaltyClient,
			orderRepo:     orderRepo,
			userRepo:      userRepo,
		}

		err := w.ProcessAccrual(context.Background(), order)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error get accrual")
	})

	t.Run("REGISTERED status -> no tx updates, nil error", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		userRepo := mocks.NewMock_UserRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)

		loyaltyClient.EXPECT().
			GetAccrual(mock.Anything, order).
			Return(models.Accrual{
				OrderNumber: "79927398713",
				Status:      "REGISTERED",
				Accrual:     0,
			}, nil).
			Once()

		w := &Worker{
			loyaltyClient: loyaltyClient,
			orderRepo:     orderRepo,
			userRepo:      userRepo,
		}

		err := w.ProcessAccrual(context.Background(), order)
		require.NoError(t, err)

		orderRepo.AssertNotCalled(t, "UpdateTx", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		userRepo.AssertNotCalled(t, "UpdateTx", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestWorker_StartWorkers_ConsumesJob(t *testing.T) {
	orderRepo := mocks.NewMock_OrderRepository(t)
	userRepo := mocks.NewMock_UserRepository(t)
	loyaltyClient := mocks.NewMock_LoyaltyClient(t)

	order := models.Order{
		ID:     1,
		Number: 79927398713,
		UserID: 10,
		Status: models.OrderStatusNew,
	}

	loyaltyClient.EXPECT().
		GetAccrual(mock.Anything, order).
		Return(models.Accrual{
			OrderNumber: "79927398713",
			Status:      "REGISTERED",
			Accrual:     0,
		}, nil).
		Once()

	w := &Worker{
		cfg: &config.Config{
			Worker: config.WorkerConfig{
				WorkerCount: 1,
			},
		},
		loyaltyClient: loyaltyClient,
		orderRepo:     orderRepo,
		userRepo:      userRepo,
		jobs:          make(chan models.Order, 1),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w.StartWorkers(ctx)
	w.jobs <- order

	time.Sleep(50 * time.Millisecond)
	cancel()
	w.wg.Wait()
}
