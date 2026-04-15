package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/handler"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type testUserHandler struct{}

func (h *testUserHandler) Register(c *gin.Context)   { c.Status(http.StatusOK) }
func (h *testUserHandler) Login(c *gin.Context)      { c.Status(http.StatusOK) }
func (h *testUserHandler) GetBalance(c *gin.Context) { c.Status(http.StatusOK) }

type testOrderHandler struct{}

func (h *testOrderHandler) Create(c *gin.Context)    { c.Status(http.StatusOK) }
func (h *testOrderHandler) GetOrders(c *gin.Context) { c.Status(http.StatusOK) }

type testWithdrawalHandler struct{}

func (h *testWithdrawalHandler) Withdrawal(c *gin.Context)     { c.Status(http.StatusOK) }
func (h *testWithdrawalHandler) GetWithdrawals(c *gin.Context) { c.Status(http.StatusOK) }

var _ handler.UserHandler = (*testUserHandler)(nil)
var _ handler.OrderHandler = (*testOrderHandler)(nil)
var _ handler.WithdrawalHandler = (*testWithdrawalHandler)(nil)

func newTestJWT() token.JWT {
	cfg := &config.Config{
		JWTToken: config.JWTTokenConfig{
			Secret:    "test-secret",
			ExpiresAt: 60,
		},
	}
	return token.NewJWT(cfg)
}

func TestSetupRouter_PublicRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	j := newTestJWT()
	r := SetupRouter(
		&testUserHandler{},
		&testOrderHandler{},
		&testWithdrawalHandler{},
		nil,
		j,
	)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "register is public",
			method:     http.MethodPost,
			path:       "/api/user/register",
			wantStatus: http.StatusOK,
		},
		{
			name:       "login is public",
			method:     http.MethodPost,
			path:       "/api/user/login",
			wantStatus: http.StatusOK,
		},
		{
			name:       "ping route exists",
			method:     http.MethodGet,
			path:       "/ping",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "unknown route -> 404",
			method:     http.MethodGet,
			path:       "/unknown",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			require.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestSetupRouter_ProtectedRoutes_RequireJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	j := newTestJWT()
	r := SetupRouter(
		&testUserHandler{},
		&testOrderHandler{},
		&testWithdrawalHandler{},
		nil,
		j,
	)

	protected := []struct {
		name   string
		method string
		path   string
	}{
		{name: "balance", method: http.MethodGet, path: "/api/user/balance"},
		{name: "orders create", method: http.MethodPost, path: "/api/user/orders"},
		{name: "orders list", method: http.MethodGet, path: "/api/user/orders"},
		{name: "withdraw", method: http.MethodPost, path: "/api/user/balance/withdraw"},
		{name: "withdrawals list", method: http.MethodGet, path: "/api/user/withdrawals"},
	}

	for _, tc := range protected {
		tc := tc
		t.Run(tc.name+" without auth -> 401", func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			require.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestSetupRouter_ProtectedRoutes_WithValidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	j := newTestJWT()
	r := SetupRouter(
		&testUserHandler{},
		&testOrderHandler{},
		&testWithdrawalHandler{},
		nil,
		j,
	)

	jwtStr, err := j.BuildJWTString(123, "max")
	require.NoError(t, err)

	protected := []struct {
		name   string
		method string
		path   string
	}{
		{name: "balance", method: http.MethodGet, path: "/api/user/balance"},
		{name: "orders create", method: http.MethodPost, path: "/api/user/orders"},
		{name: "orders list", method: http.MethodGet, path: "/api/user/orders"},
		{name: "withdraw", method: http.MethodPost, path: "/api/user/balance/withdraw"},
		{name: "withdrawals list", method: http.MethodGet, path: "/api/user/withdrawals"},
	}

	for _, tc := range protected {
		tc := tc
		t.Run(tc.name+" with valid auth -> 200", func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+jwtStr)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}
