package handler

import (
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/theplant/luhn"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// orderHandler implements OrderHandler HTTP endpoints.
type orderHandler struct {
	orderService service.OrderService
}

// NewOrderHandler creates an OrderHandler with provided order service.
func NewOrderHandler(orderService service.OrderService) OrderHandler {
	return &orderHandler{
		orderService: orderService,
	}
}

// Create handles order upload request for authenticated user.
func (h *orderHandler) Create(c *gin.Context) {
	v, exists := c.Get("userID")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	userID, ok := v.(int64)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	intNumber, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	ok = luhn.Valid(int(intNumber))
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid number"})
		return
	}

	ctx := c.Request.Context()

	if err = h.orderService.Create(ctx, intNumber, userID); err != nil {
		log.Printf("failed to create order: %v", err)
		if handleErrors(c, err) {
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusAccepted)
}

// GetOrders handles request for authenticated user's order list.
func (h *orderHandler) GetOrders(c *gin.Context) {
	v, exists := c.Get("userID")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	userID, ok := v.(int64)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}
	ctx := c.Request.Context()

	ordersResp, err := h.orderService.GetOrders(ctx, userID)
	if err != nil {
		log.Printf("failed to get orders: %v", err)
		if handleErrors(c, err) {
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, ordersResp)
}
