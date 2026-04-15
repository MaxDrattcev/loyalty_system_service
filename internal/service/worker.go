package service

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/client"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"math"
	"sync"
	"time"
)

// Worker polls accrual system and applies accrual updates to orders and user balances.
type Worker struct {
	cfg           *config.Config
	loyaltyClient client.LoyaltyClient
	jobs          chan models.Order
	wg            sync.WaitGroup
	orderRepo     repository.OrderRepository
	userRepo      repository.UserRepository
	pool          *pgxpool.Pool
}

// NewWorker creates Worker with configured dependencies and buffered jobs queue.
func NewWorker(loyaltyClient client.LoyaltyClient, cfg *config.Config, orderRepo repository.OrderRepository,
	userRepo repository.UserRepository, pool *pgxpool.Pool) *Worker {
	return &Worker{
		loyaltyClient: loyaltyClient,
		jobs:          make(chan models.Order, cfg.Worker.QueueSize),
		cfg:           cfg,
		orderRepo:     orderRepo,
		userRepo:      userRepo,
		pool:          pool,
	}
}

// Start launches worker pool and periodic reporting loop.
func (w *Worker) Start(ctx context.Context) {
	go w.StartWorkers(ctx)
	go w.StartReportingGetAccrual(ctx)
}

// StartReportingGetAccrual periodically builds snapshot of pending orders and enqueues them for processing.
func (w *Worker) StartReportingGetAccrual(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.cfg.Worker.IntervalGetAccrual) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			snapshot := w.BuildSnapshot(ctx)
			for _, order := range snapshot {
				select {
				case <-ctx.Done():
					return
				case w.jobs <- order:
				}
			}
		}
	}
}

// StartWorkers starts configured number of worker goroutines consuming jobs channel.
func (w *Worker) StartWorkers(ctx context.Context) {
	for i := 0; i < w.cfg.Worker.WorkerCount; i++ {
		w.wg.Add(1)
		go func(workerID int) {
			defer w.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case order, ok := <-w.jobs:
					if !ok {
						return
					}
					ctxWork, cancel := context.WithTimeout(ctx, 210*time.Second)
					if err := w.ProcessAccrual(ctxWork, order); err != nil {
						log.Printf("Error worker ID: %d, processing order: %v", workerID, err)
					}
					cancel()
				}
			}
		}(i)
	}
}

// ProcessAccrual fetches accrual for order and applies transactional updates to order and user.
func (w *Worker) ProcessAccrual(ctx context.Context, order models.Order) error {
	accrual, err := w.loyaltyClient.GetAccrual(ctx, order)
	if err != nil {
		return fmt.Errorf("error get accrual: %w", err)
	}
	if accrual.Status == "REGISTERED" {
		return nil
	}
	order.Accrual = int64(math.Round(accrual.Accrual * 100))
	return repository.WithTxRetry(ctx, w.pool, func(tx pgx.Tx) error {
		if err = w.orderRepo.UpdateTx(ctx, tx, accrual, order.Accrual); err != nil {
			return fmt.Errorf("error update order: %w", err)
		}
		if order.Accrual > 0 {
			if err = w.userRepo.UpdateTx(ctx, tx, order); err != nil {
				return fmt.Errorf("error update balance: %w", err)
			}
		}
		return nil
	})
}

// BuildSnapshot returns orders with NEW/PROCESSING statuses eligible for accrual polling.
func (w *Worker) BuildSnapshot(ctx context.Context) []models.Order {
	ctxWithTime, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	orders, err := w.orderRepo.GetStatusNewAndProcessing(ctxWithTime)
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		return nil
	}
	return orders
}
