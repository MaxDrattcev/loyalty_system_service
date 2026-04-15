package handler

import (
	"encoding/json"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/models"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/service"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"
)

// userHandler implements UserHandler HTTP endpoints.
type userHandler struct {
	userService service.UserService
}

// NewUserHandler creates a UserHandler with provided user service dependency.
func NewUserHandler(userService service.UserService) UserHandler {
	return &userHandler{
		userService: userService,
	}
}

// Register handles user registration request
func (u *userHandler) Register(c *gin.Context) {
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		c.Status(http.StatusUnsupportedMediaType)
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("error reading body: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Printf("error unmarshalling body: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	if !u.okLoginPassword(c, user.Login, user.Password) {
		return
	}
	ctx := c.Request.Context()
	jwt, err := u.userService.Register(ctx, user)
	if err != nil {
		log.Printf("error registering user: %v", err)
		if handleErrors(c, err) {
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Authorization", "Bearer "+jwt)
	c.Status(http.StatusOK)
}

func (u *userHandler) okLoginPassword(c *gin.Context, login string, password string) bool {
	const minLen, maxLen = 1, 100
	l := utf8.RuneCountInString(login)
	p := utf8.RuneCountInString(password)
	if login == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login is empty"})
		return false
	}
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is empty"})
		return false
	}
	if l < minLen || l > maxLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The login length is more than 100 characters"})
		return false
	}
	if p < minLen || p > maxLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The password length is more than 100 characters"})
		return false
	}
	return true
}

type errorDB struct {
	code   string
	column string
	msg    string
	status int
}

// Login handles user authentication request.
func (u *userHandler) Login(c *gin.Context) {
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		c.Status(http.StatusUnsupportedMediaType)
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("error reading body: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Printf("error unmarshalling body: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	if !u.okLoginPassword(c, user.Login, user.Password) {
		return
	}
	ctx := c.Request.Context()
	jwt, err := u.userService.Login(ctx, user)
	if err != nil {
		log.Printf("error login user: %v", err)
		if handleErrors(c, err) {
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Authorization", "Bearer "+jwt)
	c.Status(http.StatusOK)
}

// GetBalance returns authenticated user's balance.
func (u *userHandler) GetBalance(c *gin.Context) {
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

	balance, err := u.userService.GetBalance(ctx, userID)
	if err != nil {
		log.Printf("error getting balance: %v", err)
		if handleErrors(c, err) {
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, balance)
}
