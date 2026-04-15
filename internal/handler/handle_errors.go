package handler

import (
	"errors"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/repository"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func handleErrors(c *gin.Context, err error) bool {
	var dbErrors = []errorDB{
		{
			code:   "23505",
			column: "idx_users_login",
			msg:    "The login is already in use",
			status: http.StatusConflict,
		},
		{
			code:   "23505",
			column: "idx_orders_number",
			msg:    "The order number has already been uploaded by another user",
			status: http.StatusConflict,
		},
	}

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return true
	}

	var pgErr *pgconn.PgError

	for _, dbErr := range dbErrors {
		if errors.As(err, &pgErr) && pgErr.Code == dbErr.code && pgErr.ConstraintName == dbErr.column {
			c.JSON(dbErr.status, gin.H{"error": dbErr.msg})
			return true
		}
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return true
	}
	if errors.Is(err, service.ErrNoOrders) {
		c.Status(http.StatusNoContent)
		return true
	}
	if errors.Is(err, service.ErrNoWithdraws) {
		c.Status(http.StatusNoContent)
		return true
	}
	if errors.Is(err, repository.ErrInsufficientFunds) {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient funds"})
		return true
	}

	if errors.Is(err, service.ErrOrderAlreadyUploadedBySameUser) {
		c.Status(http.StatusOK)
		return true
	}

	return false
}
