package service

import (
	"context"
	"errors"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/mocks"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/token"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func newTestUserService(t *testing.T, repo *mocks.Mock_UserRepository) (UserService, token.JWT) {
	t.Helper()

	cfg := &config.Config{
		JWTToken: config.JWTTokenConfig{
			Secret:    "test-secret",
			ExpiresAt: 60,
		},
	}
	j := token.NewJWT(cfg)
	scv := NewUserService(repo, j)
	return scv, j
}

func TestUserService_Register_Success(t *testing.T) {
	repo := mocks.NewMock_UserRepository(t)
	scv, jwtToken := newTestUserService(t, repo)

	input := models.User{
		Login:    "max",
		Password: "qwerty123",
	}

	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(u models.User) bool {
			if u.Login != input.Login {
				return false
			}
			if u.Password == input.Password || u.Password == "" {
				return false
			}
			return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.Password)) == nil
		})).
		Return(models.User{
			ID:    101,
			Login: input.Login,
		}, nil).
		Once()

	jwtStr, err := scv.Register(context.Background(), input)
	require.NoError(t, err)
	require.NotEmpty(t, jwtStr)

	userID, err := jwtToken.ParseJWT(jwtStr)
	require.NoError(t, err)
	require.Equal(t, int64(101), userID)
}

func TestUserService_Register_CreateError(t *testing.T) {
	repo := mocks.NewMock_UserRepository(t)
	scv, _ := newTestUserService(t, repo)

	input := models.User{
		Login:    "max",
		Password: "qwerty123",
	}
	repoErr := errors.New("create failed")
	repo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("models.User")).
		Return(models.User{}, repoErr).
		Once()

	jwtStr, err := scv.Register(context.Background(), input)
	require.Error(t, err)
	require.ErrorIs(t, err, repoErr)
	require.Empty(t, jwtStr)
}

func TestUserService_Login_Success(t *testing.T) {
	repo := mocks.NewMock_UserRepository(t)
	scv, jwtToken := newTestUserService(t, repo)

	rawPassword := "qwerty123"
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo.EXPECT().
		GetByLogin(mock.Anything, "max").
		Return(models.User{
			ID:       7,
			Login:    "max",
			Password: string(hash),
		}, nil).
		Once()

	jwtStr, err := scv.Login(context.Background(), models.User{
		Login:    "max",
		Password: rawPassword,
	})
	require.NoError(t, err)
	require.NotEmpty(t, jwtStr)
	userID, err := jwtToken.ParseJWT(jwtStr)
	require.NoError(t, err)
	require.Equal(t, int64(7), userID)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	repo := mocks.NewMock_UserRepository(t)
	scv, _ := newTestUserService(t, repo)

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo.EXPECT().
		GetByLogin(mock.Anything, "max").
		Return(models.User{
			ID:       7,
			Login:    "max",
			Password: string(hash),
		}, nil).
		Once()

	jwtStr, err := scv.Login(context.Background(), models.User{
		Login:    "max",
		Password: "wrong-password",
	})
	require.Error(t, err)
	require.Empty(t, jwtStr)
	require.Contains(t, err.Error(), "invalid password")
}

func TestUserService_GetBalance_Success(t *testing.T) {
	repo := mocks.NewMock_UserRepository(t)
	scv, _ := newTestUserService(t, repo)

	repo.EXPECT().
		GetByUserID(mock.Anything, int64(1)).
		Return(models.User{
			ID:             1,
			CurrentBalance: 1234,
			TotalWithdrawn: 200,
		}, nil).
		Once()

	got, err := scv.GetBalance(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, models.BalanceResponse{
		Current:   12.34,
		Withdrawn: 2.00,
	}, got)
}

func TestUserService_GetBalance_RepoError(t *testing.T) {
	repo := mocks.NewMock_UserRepository(t)
	svc, _ := newTestUserService(t, repo)
	repoErr := errors.New("db unavailable")
	repo.EXPECT().
		GetByUserID(mock.Anything, int64(1)).
		Return(models.User{}, repoErr).
		Once()
	got, err := svc.GetBalance(context.Background(), 1)
	require.Error(t, err)
	require.ErrorIs(t, err, repoErr)
	require.Equal(t, models.BalanceResponse{}, got)
}
