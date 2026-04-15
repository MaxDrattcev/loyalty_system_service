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
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRequest := func(h OrderHandler, withUserID bool, userID any, body string) *httptest.ResponseRecorder {
		router := gin.New()
		if withUserID {
			router.Use(func(c *gin.Context) {
				c.Set("userID", userID)
				c.Next()
			})
		}
		router.POST("/api/user/orders", h.Create)

		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		return w
	}

	t.Run("success: 202", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		orderNumber := int64(79927398713)

		mockSvc.EXPECT().
			Create(mock.Anything, orderNumber, int64(1)).
			Return(nil).
			Once()

		w := makeRequest(h, true, int64(1), "79927398713")
		require.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("no userID in context: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		w := makeRequest(h, false, nil, "79927398713")
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("userID wrong type: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		w := makeRequest(h, true, "not-int64", "79927398713")
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("empty body: 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		w := makeRequest(h, true, int64(1), "   ")
		require.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("not a number: 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		w := makeRequest(h, true, int64(1), "abc")
		require.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("invalid luhn number: 422", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		w := makeRequest(h, true, int64(1), "1234567890")
		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("order already uploaded by same user: 200", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		orderNumber := int64(79927398713)
		mockSvc.EXPECT().
			Create(mock.Anything, orderNumber, int64(1)).
			Return(service.ErrOrderAlreadyUploadedBySameUser).
			Once()

		w := makeRequest(h, true, int64(1), "79927398713")
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service returns unknown error: 500", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		orderNumber := int64(79927398713)
		mockSvc.EXPECT().
			Create(mock.Anything, orderNumber, int64(1)).
			Return(errors.New("service error")).
			Once()

		w := makeRequest(h, true, int64(1), "79927398713")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestOrderHandler_GetOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRequest := func(h OrderHandler, withUserID bool, userID any) *httptest.ResponseRecorder {
		router := gin.New()
		if withUserID {
			router.Use(func(c *gin.Context) {
				c.Set("userID", userID)
				c.Next()
			})
		}
		router.GET("/api/user/orders", h.GetOrders)

		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		return w
	}

	t.Run("success: 200 + json", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		accrual := 12.34
		output := []models.OrderResponse{
			{
				Number:   "79927398713",
				Status:   models.OrderStatusProcessed,
				Accrual:  &accrual,
				Uploaded: time.Now().UTC().Format(time.RFC3339),
			},
		}

		mockSvc.EXPECT().
			GetOrders(mock.Anything, int64(1)).
			Return(output, nil).
			Once()

		w := makeRequest(h, true, int64(1))
		require.Equal(t, http.StatusOK, w.Code)

		var got []models.OrderResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		require.NoError(t, err)
		require.Equal(t, output, got)
	})

	t.Run("no userID in context: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		w := makeRequest(h, false, nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GetOrders", mock.Anything, mock.Anything)
	})

	t.Run("userID wrong type: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		w := makeRequest(h, true, "not-int64")
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GetOrders", mock.Anything, mock.Anything)
	})

	t.Run("no orders: 204", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		mockSvc.EXPECT().
			GetOrders(mock.Anything, int64(1)).
			Return([]models.OrderResponse(nil), service.ErrNoOrders).
			Once()

		w := makeRequest(h, true, int64(1))
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("service returns unknown error: 500", func(t *testing.T) {
		mockSvc := &mocks.Mock_OrderService{}
		h := NewOrderHandler(mockSvc)

		mockSvc.EXPECT().
			GetOrders(mock.Anything, int64(1)).
			Return([]models.OrderResponse(nil), errors.New("service error")).
			Once()

		w := makeRequest(h, true, int64(1))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
