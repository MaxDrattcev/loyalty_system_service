package handler

import "github.com/gin-gonic/gin"

// UserHandler handles HTTP requests for user operations.
type UserHandler interface {
	// Register validates request payload, creates a new user, and sets JWT in Authorization header.
	// Returns:
	// - 200 on success
	// - 400 on invalid payload/validation error
	// - 415 on unsupported content type
	// - 409 when login already exists
	// - 500 on unexpected internal error
	Register(c *gin.Context)

	// Login authenticates user credentials and sets JWT in Authorization header.
	// Returns:
	// - 200 on success
	// - 400 on invalid payload/validation error
	// - 415 on unsupported content type
	// - 401 on invalid credentials
	// - 500 on unexpected internal error
	Login(c *gin.Context)

	// GetBalance returns current user balance and total withdrawn amount.
	// Requires authenticated userID in request context (set by JWT middleware).
	// Returns:
	// - 200 with JSON body on success
	// - 401 when user is unauthorized
	// - 500 on unexpected internal error
	GetBalance(c *gin.Context)
}

// OrderHandler handles HTTP requests for order operations.
type OrderHandler interface {
	// Create accepts user order number, validates it, and schedules accrual processing.
	// Requires authenticated userID in request context.
	// Returns:
	// - 202 when order is accepted
	// - 200 when order was already uploaded by same user
	// - 400 on invalid request body
	// - 401 when user is unauthorized
	// - 422 when order number is invalid (Luhn)
	// - 409 when order belongs to another user
	// - 500 on unexpected internal error
	Create(c *gin.Context)

	// GetOrders returns all uploaded orders for authenticated user.
	// Requires authenticated userID in request context.
	// Returns:
	// - 200 with JSON body on success
	// - 204 when user has no orders
	// - 401 when user is unauthorized
	// - 500 on unexpected internal error
	GetOrders(c *gin.Context)
}

// WithdrawalHandler handles HTTP requests for withdrawal operations.
type WithdrawalHandler interface {
	// Withdrawal handles withdrawal request for authenticated user.
	// Requires authenticated userID in request context.
	// Returns:
	// - 200 on success
	// - 400 on invalid payload or non-positive sum
	// - 401 when user is unauthorized
	// - 402 when user has insufficient funds
	// - 415 on unsupported content type
	// - 422 when order number is invalid (Luhn)
	// - 500 on unexpected internal error
	Withdrawal(c *gin.Context)

	// GetWithdrawals returns withdrawal history for authenticated user.
	// Requires authenticated userID in request context.
	// Returns:
	// - 200 with JSON body on success
	// - 204 when no withdrawals exist
	// - 401 when user is unauthorized
	// - 500 on unexpected internal error
	GetWithdrawals(c *gin.Context)
}
