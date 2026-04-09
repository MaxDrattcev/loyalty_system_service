package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/client"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

var (
	ErrNoOrders                       = errors.New("no orders")
	ErrOrderAlreadyUploadedBySameUser = errors.New("order already uploaded by same user")
)

type orderService struct {
	orderRepo     repository.OrderRepository
	loyaltyClient client.LoyaltyClient
	worker        *Worker
}

func NewOrderService(orderRepo repository.OrderRepository, loyaltyClient client.LoyaltyClient, worker *Worker) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		loyaltyClient: loyaltyClient,
		worker:        worker,
	}
}

func (s *orderService) Create(ctx context.Context, numberOrder, userID int64) error {
	order := models.Order{
		Number: numberOrder,
		UserID: userID,
		Status: models.OrderStatusNew,
	}
	var pgErr *pgconn.PgError

	err := s.orderRepo.Create(ctx, order)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			existOrder, err1 := s.orderRepo.GetByNumber(ctx, numberOrder)
			if err1 != nil {
				return err1
			}
			if existOrder.UserID != userID {
				return fmt.Errorf("the order number: %d has already been uploaded by another user: %w", numberOrder, err)
			} else {
				return fmt.Errorf("the order number: %d has already been uploaded: %w", numberOrder, ErrOrderAlreadyUploadedBySameUser)
			}
		}
		return err
	}
	s.worker.jobs <- order
	return nil
}

func (s *orderService) GetOrders(ctx context.Context, userID int64) ([]models.OrderResponse, error) {
	orders, err := s.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders found for user %d: %w", userID, ErrNoOrders)
	}
	var ordersResp []models.OrderResponse

	for _, order := range orders {
		var orderResp models.OrderResponse
		strOrder := fmt.Sprint(order.Number)
		orderResp.Number = strOrder
		orderResp.Status = order.Status
		if order.Accrual > 0 {
			floatAccrual := float64(order.Accrual) / 100
			orderResp.Accrual = &floatAccrual
		}
		orderResp.Uploaded = order.UploadedAt.Format(time.RFC3339)

		ordersResp = append(ordersResp, orderResp)
	}
	return ordersResp, nil
}
