package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/mocks"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRequest := func(h UserHandler, body []byte, contentType string) *httptest.ResponseRecorder {
		router := gin.New()
		router.POST("/api/user/register", h.Register)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("success: 200 + Authorization header", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)

		input := models.User{
			Login:    "max",
			Password: "qwerty123",
		}

		mockSvc.EXPECT().
			Register(mock.Anything, input).
			Return("jwt-token-123", nil).
			Once()

		payload, _ := json.Marshal(input)
		w := makeRequest(h, payload, "application/json")

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unsupported content-type: 415", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)

		w := makeRequest(h, []byte(`{"login":"a","password":"b"}`), "text/plain")

		require.Equal(t, http.StatusUnsupportedMediaType, w.Code)
		mockSvc.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	})

	t.Run("invalid json: 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, []byte(`{"login":`), "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	})

	t.Run("validation failed (empty login): 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, []byte(`{"login":"","password":"abc"}`), "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	})

	t.Run("validation failed (empty password): 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, []byte(`{"login":"max","password":""}`), "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service returns error: 500", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		input := models.User{
			Login:    "max",
			Password: "qwerty123",
		}
		mockSvc.EXPECT().
			Register(mock.Anything, input).
			Return("", errors.New("service error")).
			Once()

		payload, _ := json.Marshal(input)
		w := makeRequest(h, payload, "application/json")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

}

func TestUserHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRequest := func(h UserHandler, body []byte, contentType string) *httptest.ResponseRecorder {
		router := gin.New()
		router.POST("/api/user/login", h.Login)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("success: 200 + Authorization header", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		input := models.User{
			Login:    "max",
			Password: "qwerty123",
		}
		mockSvc.EXPECT().
			Login(mock.Anything, input).
			Return("jwt-token-123", nil).
			Once()

		payload, _ := json.Marshal(input)
		w := makeRequest(h, payload, "application/json")
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unsupported content-type: 415", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, []byte(`{"login":"a","password":"b"}`), "text/plain")
		require.Equal(t, http.StatusUnsupportedMediaType, w.Code)
	})

	t.Run("invalid json: 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, []byte(`{"login":"a"}`), "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation failed (empty login): 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, []byte(`{"login":"","password":"b"}`), "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation failed (empty password): 400", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, []byte(`{"login":"a","password":""}`), "application/json")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service returns error: 500", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		input := models.User{
			Login:    "max",
			Password: "qwerty123",
		}
		mockSvc.EXPECT().
			Login(mock.Anything, input).
			Return("", errors.New("service error")).
			Once()
		payload, _ := json.Marshal(input)
		w := makeRequest(h, payload, "application/json")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

}

func TestUserHandler_GetBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRequest := func(h UserHandler, withUserID bool, userID any) *httptest.ResponseRecorder {
		router := gin.New()
		if withUserID {
			router.Use(func(c *gin.Context) {
				c.Set("userID", userID)
				c.Next()
			})
		}
		router.GET("/api/user/balance", h.GetBalance)

		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	t.Run("success: 200 + JSON balance", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		output := models.BalanceResponse{
			Current:   12.12,
			Withdrawn: 5.55,
		}
		mockSvc.EXPECT().
			GetBalance(mock.Anything, int64(1)).
			Return(output, nil).
			Once()

		w := makeRequest(h, true, int64(1))

		require.Equal(t, http.StatusOK, w.Code)

		var got models.BalanceResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		require.NoError(t, err)
	})

	t.Run("no userID in context: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, false, nil)
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GetBalance", mock.Anything, mock.Anything)
	})

	t.Run("userID wrong type: 401", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		w := makeRequest(h, true, "not-int64")
		require.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GetBalance", mock.Anything, mock.Anything)
	})

	t.Run("service returns error: 500", func(t *testing.T) {
		mockSvc := &mocks.Mock_UserService{}
		h := NewUserHandler(mockSvc)
		mockSvc.EXPECT().
			GetBalance(mock.Anything, int64(1)).
			Return(models.BalanceResponse{}, errors.New("service error")).
			Once()
		w := makeRequest(h, true, int64(1))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
