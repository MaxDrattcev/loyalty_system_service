package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/mocks"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success: creates order and pushes job to worker channel", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)

		worker := &Worker{jobs: make(chan models.Order, 1)}
		svc := NewOrderService(orderRepo, loyaltyClient, worker)

		numberOrder := int64(79927398713)
		userID := int64(10)

		orderRepo.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(o models.Order) bool {
				return o.Number == numberOrder &&
					o.UserID == userID &&
					o.Status == models.OrderStatusNew
			})).
			Return(nil).
			Once()

		err := svc.Create(ctx, numberOrder, userID)
		require.NoError(t, err)

		select {
		case got := <-worker.jobs:
			require.Equal(t, numberOrder, got.Number)
			require.Equal(t, userID, got.UserID)
			require.Equal(t, models.OrderStatusNew, got.Status)
		default:
			t.Fatal("expected order to be pushed into worker.jobs")
		}
	})

	t.Run("duplicate order by same user -> ErrOrderAlreadyUploadedBySameUser", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)
		worker := &Worker{jobs: make(chan models.Order, 1)}
		svc := NewOrderService(orderRepo, loyaltyClient, worker)

		numberOrder := int64(79927398713)
		userID := int64(10)

		pgErr := &pgconn.PgError{Code: "23505"}

		orderRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("models.Order")).
			Return(pgErr).
			Once()

		orderRepo.EXPECT().
			GetByNumber(mock.Anything, numberOrder).
			Return(models.Order{Number: numberOrder, UserID: userID}, nil).
			Once()

		err := svc.Create(ctx, numberOrder, userID)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrOrderAlreadyUploadedBySameUser)
	})

	t.Run("duplicate order by another user -> conflict error", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)
		worker := &Worker{jobs: make(chan models.Order, 1)}
		svc := NewOrderService(orderRepo, loyaltyClient, worker)

		numberOrder := int64(79927398713)
		userID := int64(10)

		pgErr := &pgconn.PgError{Code: "23505"}

		orderRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("models.Order")).
			Return(pgErr).
			Once()

		orderRepo.EXPECT().
			GetByNumber(mock.Anything, numberOrder).
			Return(models.Order{Number: numberOrder, UserID: 999}, nil).
			Once()

		err := svc.Create(ctx, numberOrder, userID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already been uploaded by another user")
		require.ErrorIs(t, err, pgErr)
	})

	t.Run("duplicate create and getByNumber failed -> returns getByNumber error", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)
		worker := &Worker{jobs: make(chan models.Order, 1)}
		svc := NewOrderService(orderRepo, loyaltyClient, worker)

		numberOrder := int64(79927398713)
		userID := int64(10)

		pgErr := &pgconn.PgError{Code: "23505"}
		getErr := errors.New("get by number failed")

		orderRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("models.Order")).
			Return(pgErr).
			Once()

		orderRepo.EXPECT().
			GetByNumber(mock.Anything, numberOrder).
			Return(models.Order{}, getErr).
			Once()

		err := svc.Create(ctx, numberOrder, userID)
		require.Error(t, err)
		require.ErrorIs(t, err, getErr)
	})

	t.Run("non-duplicate create error -> returns as is", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)
		worker := &Worker{jobs: make(chan models.Order, 1)}
		svc := NewOrderService(orderRepo, loyaltyClient, worker)

		numberOrder := int64(79927398713)
		userID := int64(10)

		createErr := errors.New("db unavailable")

		orderRepo.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("models.Order")).
			Return(createErr).
			Once()

		err := svc.Create(ctx, numberOrder, userID)
		require.Error(t, err)
		require.ErrorIs(t, err, createErr)
	})
}

func TestOrderService_GetOrders(t *testing.T) {
	ctx := context.Background()

	t.Run("success: maps domain orders to response", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)
		worker := &Worker{jobs: make(chan models.Order, 1)}
		svc := NewOrderService(orderRepo, loyaltyClient, worker)

		userID := int64(5)
		uploaded := time.Date(2026, 4, 15, 10, 30, 0, 0, time.UTC)

		orderRepo.EXPECT().
			GetOrdersByUserID(mock.Anything, userID).
			Return([]models.Order{
				{
					Number:     79927398713,
					Status:     models.OrderStatusProcessed,
					Accrual:    1234,
					UploadedAt: uploaded,
				},
				{
					Number:     11111111111,
					Status:     models.OrderStatusNew,
					Accrual:    0,
					UploadedAt: uploaded,
				},
			}, nil).
			Once()

		got, err := svc.GetOrders(ctx, userID)
		require.NoError(t, err)
		require.Len(t, got, 2)

		require.Equal(t, "79927398713", got[0].Number)
		require.Equal(t, models.OrderStatusProcessed, got[0].Status)
		require.NotNil(t, got[0].Accrual)
		require.Equal(t, 12.34, *got[0].Accrual)
		require.Equal(t, uploaded.Format(time.RFC3339), got[0].Uploaded)

		require.Equal(t, "11111111111", got[1].Number)
		require.Equal(t, models.OrderStatusNew, got[1].Status)
		require.Nil(t, got[1].Accrual)
		require.Equal(t, uploaded.Format(time.RFC3339), got[1].Uploaded)
	})

	t.Run("repo error -> returns error", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)
		worker := &Worker{jobs: make(chan models.Order, 1)}
		svc := NewOrderService(orderRepo, loyaltyClient, worker)

		userID := int64(5)
		repoErr := errors.New("repo failed")

		orderRepo.EXPECT().
			GetOrdersByUserID(mock.Anything, userID).
			Return(nil, repoErr).
			Once()

		got, err := svc.GetOrders(ctx, userID)
		require.Error(t, err)
		require.ErrorIs(t, err, repoErr)
		require.Nil(t, got)
	})

	t.Run("empty orders -> ErrNoOrders", func(t *testing.T) {
		orderRepo := mocks.NewMock_OrderRepository(t)
		loyaltyClient := mocks.NewMock_LoyaltyClient(t)
		worker := &Worker{jobs: make(chan models.Order, 1)}
		svc := NewOrderService(orderRepo, loyaltyClient, worker)

		userID := int64(5)

		orderRepo.EXPECT().
			GetOrdersByUserID(mock.Anything, userID).
			Return([]models.Order{}, nil).
			Once()

		got, err := svc.GetOrders(ctx, userID)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNoOrders)
		require.Nil(t, got)
	})
}
