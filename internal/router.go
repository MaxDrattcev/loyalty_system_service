package internal

import (
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/handler"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/middleware"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

// SetupRouter configures HTTP routes and middleware for all application endpoints.
func SetupRouter(userHandler handler.UserHandler, orderHandler handler.OrderHandler, withdrawalHandler handler.WithdrawalHandler,
	pool *pgxpool.Pool, jwt token.JWT) http.Handler {
	router := gin.New()

	router.Use(gin.Recovery())

	router.Use(middleware.Logger(), middleware.Compress())

	router.POST("/api/user/register", userHandler.Register)
	router.POST("/api/user/login", userHandler.Login)
	router.GET("/api/user/balance", middleware.CheckJWT(&jwt), userHandler.GetBalance)

	router.POST("/api/user/orders", middleware.CheckJWT(&jwt), orderHandler.Create)
	router.GET("/api/user/orders", middleware.CheckJWT(&jwt), orderHandler.GetOrders)

	router.POST("/api/user/balance/withdraw", middleware.CheckJWT(&jwt), withdrawalHandler.Withdrawal)
	router.GET("/api/user/withdrawals", middleware.CheckJWT(&jwt), withdrawalHandler.GetWithdrawals)

	router.GET("/ping", handler.PingDB(pool))

	return router
}
