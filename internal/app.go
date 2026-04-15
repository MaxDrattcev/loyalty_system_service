package internal

import (
	"context"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/client"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/handler"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/repository"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/service"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/token"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
)

// App represents assembled application with HTTP router and runtime config.
type App struct {
	userHandler handler.UserHandler
	cfg         *config.Config
	router      http.Handler
}

// NewApp wires dependencies, starts background worker, and returns initialized App.
func NewApp(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool) *App {
	jwt := token.NewJWT(cfg)

	loyaltyClient := client.NewLoyaltyClient(cfg)

	userRepo := repository.NewUserRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	withdrawalRepo := repository.NewWithdrawalRepository(pool)

	worker := service.NewWorker(loyaltyClient, cfg, orderRepo, userRepo, pool)
	worker.Start(ctx)

	userService := service.NewUserService(userRepo, jwt)
	orderService := service.NewOrderService(orderRepo, loyaltyClient, worker)
	withdrawalService := service.NewWithdrawalService(orderRepo, userRepo, withdrawalRepo, pool)

	userHandler := handler.NewUserHandler(userService)
	orderHandler := handler.NewOrderHandler(orderService)
	withdrawalHandler := handler.NewWithdrawalHandler(withdrawalService)

	router := SetupRouter(userHandler, orderHandler, withdrawalHandler, pool, jwt)
	return &App{
		userHandler: userHandler,
		cfg:         cfg,
		router:      router,
	}
}

// Run starts HTTP server with configured address and router.
func (a *App) Run() error {
	log.Printf("Starting server on %s", a.cfg.Server.Address)
	return http.ListenAndServe(a.cfg.Server.Address, a.router)
}
