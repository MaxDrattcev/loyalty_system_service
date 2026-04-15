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

// userService implements user business logic.
type userService struct {
	userRepository repository.UserRepository
	jwt            token.JWT
	cfg            *config.Config
}

// NewUserService creates a UserService with user repository and JWT provider.
func NewUserService(userRepository repository.UserRepository, jwt token.JWT) UserService {
	return &userService{
		userRepository: userRepository,
		jwt:            jwt,
	}
}

// Register hashes user password, stores user in repository, and returns JWT token.
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

// hashPassword returns bcrypt hash of plain password.
func (u *userService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Login verifies credentials and returns JWT token for an existing user.
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

// checkPassword compares plain password with bcrypt hash from storage.
func (u *userService) checkPassword(password, hashFromDB string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashFromDB), []byte(password))
}

// GetBalance returns user's current and withdrawn balances converted from cents to float values.
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
