package service

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/repository"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/token"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepository repository.UserRepository
	jwt            token.JWT
	cfg            *config.Config
}

func NewUserService(userRepository repository.UserRepository, jwt token.JWT) UserService {
	return &userService{
		userRepository: userRepository,
		jwt:            jwt,
	}
}

func (u *userService) Register(ctx context.Context, user models.User) (string, error) {
	hash, err := u.hashPassword(user.Password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hash
	savedUser, err := u.userRepository.Create(ctx, user)
	if err != nil {
		return "", err
	}
	return u.jwt.BuildJWTString(savedUser.ID, savedUser.Login)
}

func (u *userService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (u *userService) Login(ctx context.Context, user models.User) (string, error) {
	existingUser, err := u.userRepository.GetByLogin(ctx, user.Login)
	if err != nil {
		return "", err
	}
	if err := u.checkPassword(user.Password, existingUser.Password); err != nil {
		return "", fmt.Errorf("invalid password for user id %d: %w", existingUser.ID, err)
	}
	return u.jwt.BuildJWTString(existingUser.ID, existingUser.Login)
}

func (u *userService) checkPassword(password, hashFromDB string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashFromDB), []byte(password))
}

func (u *userService) GetBalance(ctx context.Context, userID int64) (models.BalanceResponse, error) {

	user, err := u.userRepository.GetByUserID(ctx, userID)
	if err != nil {
		return models.BalanceResponse{}, err
	}
	var balanceResp = models.BalanceResponse{
		Current:   float64(user.CurrentBalance) / 100,
		Withdrawn: float64(user.TotalWithdrawn) / 100,
	}
	return balanceResp, nil
}
