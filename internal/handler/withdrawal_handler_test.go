package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/mocks"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/repository"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalHandler_Withdrawal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRequest := func(h WithdrawalHandler, withUserID bool, userID any, body string, contentType string) *httptest.ResponseRecorder {
		router := gin.New()
		if withUserID {
			router.Use(func(c *gin.Context) {
				c.Set("userID", userID)
				c.Next()
			})
		}
		router.POST("/api/user/balance/withdraw", h.Withdrawal)

		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(body))
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("success: 200", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		// Валидный Luhn-номер + положительная сумма
		body := `{"order":"79927398713","sum":12.34}`

		mockSvc.EXPECT().
			Withdrawal(mock.Anything, int64(1), int64(79927398713), 12.34).
			Return(nil).
			Once()

		w := makeRequest(h, true, int64(1), body, "application/json")
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unsupported content-type: 415", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, true, int64(1), `{"order":"79927398713","sum":12.34}`, "text/plain")
		require.Equal(t, http.StatusUnsupportedMediaType, w.Code)
		mockSvc.AssertNotCalled(t, "Withdrawal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("no userID in context: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, false, nil, `{"order":"79927398713","sum":12.34}`, "application/json")
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "Withdrawal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("userID wrong type: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, true, "not-int64", `{"order":"79927398713","sum":12.34}`, "application/json")
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "Withdrawal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("invalid json: 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, true, int64(1), `{"order":`, "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "Withdrawal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("order is not numeric: 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, true, int64(1), `{"order":"abc","sum":12.34}`, "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "Withdrawal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("invalid luhn number: 422", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, true, int64(1), `{"order":"1234567890","sum":12.34}`, "application/json")
		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockSvc.AssertNotCalled(t, "Withdrawal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("sum <= 0: 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, true, int64(1), `{"order":"79927398713","sum":0}`, "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "Withdrawal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("insufficient funds: 402", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		mockSvc.EXPECT().
			Withdrawal(mock.Anything, int64(1), int64(79927398713), 12.34).
			Return(repository.ErrInsufficientFunds).
			Once()

		w := makeRequest(h, true, int64(1), `{"order":"79927398713","sum":12.34}`, "application/json")
		require.Equal(t, http.StatusPaymentRequired, w.Code)
	})

	t.Run("service returns unknown error: 500", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		mockSvc.EXPECT().
			Withdrawal(mock.Anything, int64(1), int64(79927398713), 12.34).
			Return(errors.New("service error")).
			Once()

		w := makeRequest(h, true, int64(1), `{"order":"79927398713","sum":12.34}`, "application/json")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestWithdrawalHandler_GetWithdrawals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRequest := func(h WithdrawalHandler, withUserID bool, userID any) *httptest.ResponseRecorder {
		router := gin.New()
		if withUserID {
			router.Use(func(c *gin.Context) {
				c.Set("userID", userID)
				c.Next()
			})
		}
		router.GET("/api/user/withdrawals", h.GetWithdrawals)

		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("success: 200 + json", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		processedAt := time.Now().UTC().Format(time.RFC3339)
		output := []models.WithdrawalResponse{
			{
				Order:       "79927398713",
				Sum:         12.34,
				ProcessedAt: &processedAt,
			},
		}

		mockSvc.EXPECT().
			GetWithdrawals(mock.Anything, int64(1)).
			Return(output, nil).
			Once()

		w := makeRequest(h, true, int64(1))
		require.Equal(t, http.StatusOK, w.Code)

		var got []models.WithdrawalResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		require.NoError(t, err)
		require.Equal(t, output, got)
	})

	t.Run("no userID in context: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, false, nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GetWithdrawals", mock.Anything, mock.Anything)
	})

	t.Run("userID wrong type: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		w := makeRequest(h, true, "not-int64")
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GetWithdrawals", mock.Anything, mock.Anything)
	})

	t.Run("no withdrawals: 204", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		mockSvc.EXPECT().
			GetWithdrawals(mock.Anything, int64(1)).
			Return([]models.WithdrawalResponse{}, service.ErrNoWithdraws).
			Once()

		w := makeRequest(h, true, int64(1))
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("service returns unknown error: 500", func(t *testing.T) {
		mockSvc := &mocks.Mock_WithdrawalService{}
		h := NewWithdrawalHandler(mockSvc)

		mockSvc.EXPECT().
			GetWithdrawals(mock.Anything, int64(1)).
			Return([]models.WithdrawalResponse(nil), errors.New("service error")).
			Once()

		w := makeRequest(h, true, int64(1))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
