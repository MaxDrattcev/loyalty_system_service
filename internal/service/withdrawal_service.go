package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"math"
	"time"
)

// ErrNoWithdraws indicates that user has no withdrawal history.
var ErrNoWithdraws = errors.New("no withdraws")

// withdrawalService implements withdrawal business logic.
type withdrawalService struct {
	orderRepo      repository.OrderRepository
	userRepo       repository.UserRepository
	withdrawalRepo repository.WithdrawalRepository
	pool           *pgxpool.Pool
}

// NewWithdrawalService creates WithdrawalService with required repositories and DB pool.
func NewWithdrawalService(orderRepo repository.OrderRepository, userRepo repository.UserRepository,
	withdrawalRepo repository.WithdrawalRepository, pool *pgxpool.Pool) WithdrawalService {
	return &withdrawalService{
		orderRepo:      orderRepo,
		userRepo:       userRepo,
		withdrawalRepo: withdrawalRepo,
		pool:           pool,
	}
}

// Withdrawal performs atomic withdrawal transaction:
// creates order record, decreases user balance (if enough funds), and stores withdrawal history.
func (w *withdrawalService) Withdrawal(ctx context.Context, userID, order int64, sum float64) error {
	intWd := int64(math.Round(sum * 100))
	return repository.WithTxRetry(ctx, w.pool, func(tx pgx.Tx) error {

		orderStr := models.Order{
			Number: order,
			UserID: userID,
			Status: models.OrderStatusProcessed,
		}
		if err := w.orderRepo.CreateTx(ctx, tx, orderStr); err != nil {
			return err
		}
		if err := w.userRepo.WithdrawIfEnoughTx(ctx, tx, userID, intWd); err != nil {
			return err
		}

		wh := models.WithdrawalHistory{
			UserID:        userID,
			OrderNumber:   order,
			SumWithdrawal: intWd,
		}
		if err := w.withdrawalRepo.CreateTx(ctx, tx, wh); err != nil {
			return err
		}
		return nil
	})
}

// GetWithdrawals returns user withdrawals in API response format.
// Converts stored cents to float values and formats processed time as RFC3339.
func (w *withdrawalService) GetWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalResponse, error) {
	var withdrawalsResp []models.WithdrawalResponse
	withdrawals, err := w.withdrawalRepo.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting withdrawals for user %d: %w", userID, err)
	}
	if len(withdrawals) == 0 {
		return []models.WithdrawalResponse{}, fmt.Errorf("%w: %d", ErrNoWithdraws, userID)
	}

	for _, wd := range withdrawals {
		var withdrawalResp models.WithdrawalResponse
		withdrawalResp.Order = fmt.Sprint(wd.OrderNumber)
		floatSum := float64(wd.SumWithdrawal) / 100
		withdrawalResp.Sum = floatSum
		stringTime := wd.ProcessedAt.Format(time.RFC3339)
		withdrawalResp.ProcessedAt = &stringTime

		withdrawalsResp = append(withdrawalsResp, withdrawalResp)
	}
	return withdrawalsResp, nil
}
