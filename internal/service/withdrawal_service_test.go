package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/mocks"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalService_GetWithdrawals(t *testing.T) {
	ctx := context.Background()

	t.Run("success: maps history to response", func(t *testing.T) {
		withdrawalRepo := mocks.NewMock_WithdrawalRepository(t)

		svc := &withdrawalService{
			withdrawalRepo: withdrawalRepo,
		}

		userID := int64(10)
		processedAt := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)

		withdrawalRepo.EXPECT().
			GetWithdrawals(mock.Anything, userID).
			Return([]models.WithdrawalHistory{
				{
					UserID:        userID,
					OrderNumber:   79927398713,
					SumWithdrawal: 1234,
					ProcessedAt:   processedAt,
				},
			}, nil).
			Once()

		got, err := svc.GetWithdrawals(ctx, userID)
		require.NoError(t, err)
		require.Len(t, got, 1)

		require.Equal(t, "79927398713", got[0].Order)
		require.Equal(t, 12.34, got[0].Sum)
		require.NotNil(t, got[0].ProcessedAt)
		require.Equal(t, processedAt.Format(time.RFC3339), *got[0].ProcessedAt)
	})

	t.Run("repo error: wrapped error with user id", func(t *testing.T) {
		withdrawalRepo := mocks.NewMock_WithdrawalRepository(t)
		svc := &withdrawalService{
			withdrawalRepo: withdrawalRepo,
		}

		userID := int64(10)
		repoErr := errors.New("db unavailable")

		withdrawalRepo.EXPECT().
			GetWithdrawals(mock.Anything, userID).
			Return(nil, repoErr).
			Once()

		got, err := svc.GetWithdrawals(ctx, userID)
		require.Error(t, err)
		require.ErrorIs(t, err, repoErr)
		require.Contains(t, err.Error(), "error getting withdrawals for user 10")
		require.Nil(t, got)
	})

	t.Run("empty list: ErrNoWithdraws", func(t *testing.T) {
		withdrawalRepo := mocks.NewMock_WithdrawalRepository(t)
		svc := &withdrawalService{
			withdrawalRepo: withdrawalRepo,
		}

		userID := int64(10)

		withdrawalRepo.EXPECT().
			GetWithdrawals(mock.Anything, userID).
			Return([]models.WithdrawalHistory{}, nil).
			Once()

		got, err := svc.GetWithdrawals(ctx, userID)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNoWithdraws)

		require.NotNil(t, got)
		require.Len(t, got, 0)
	})
}
