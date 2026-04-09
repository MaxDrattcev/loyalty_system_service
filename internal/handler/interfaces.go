package handler

import "github.com/gin-gonic/gin"

type UserHandler interface {
	Register(c *gin.Context)

	Login(c *gin.Context)

	GetBalance(c *gin.Context)
}

type OrderHandler interface {
	Create(c *gin.Context)

	GetOrders(c *gin.Context)
}

type WithdrawalHandler interface {
	Withdrawal(c *gin.Context)

	GetWithdrawals(c *gin.Context)
}
