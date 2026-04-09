package handler

import (
	"encoding/json"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/theplant/luhn"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type withdrawalHandler struct {
	withdrawalService service.WithdrawalService
}

func NewWithdrawalHandler(withdrawalService service.WithdrawalService) WithdrawalHandler {
	return &withdrawalHandler{
		withdrawalService: withdrawalService,
	}
}

func (h *withdrawalHandler) Withdrawal(c *gin.Context) {
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		c.Status(http.StatusUnsupportedMediaType)
		return
	}

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
		log.Printf("error reading body: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	var withdrawal models.Withdrawal
	if err := json.Unmarshal(body, &withdrawal); err != nil {
		log.Printf("error unmarshalling body: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	intOrder, err := strconv.ParseInt(withdrawal.Order, 10, 64)

	if err != nil {
		log.Printf("error parsing order: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	ok = luhn.Valid(int(intOrder))
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid number"})
		return
	}
	if withdrawal.Sum <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sum must be positive"})
		return
	}
	ctx := c.Request.Context()
	if err := h.withdrawalService.Withdrawal(ctx, userID, intOrder, withdrawal.Sum); err != nil {
		log.Println("error withdrawal:", err)
		if handleErrors(c, err) {
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (h *withdrawalHandler) GetWithdrawals(c *gin.Context) {
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
	withdrawals, err := h.withdrawalService.GetWithdrawals(ctx, userID)
	if err != nil {
		log.Println("error getting withdrawals:", err)
		if handleErrors(c, err) {
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, withdrawals)
}
