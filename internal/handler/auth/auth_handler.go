package auth

import (
	"errors"
	"time"

	"go-short/internal/service"
	"go-short/internal/util"

	"github.com/gin-gonic/gin"
)

// AuthHandler 负责认证相关接口：注册、登录
type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}
	registerUserCommand := service.RegisterUserCommand{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}
	user, err := h.userService.CreateUser(c, registerUserCommand)
	if err != nil {
		// 检查是否是用户名已存在错误
		if errors.Is(err, service.ErrUserExists) {
			c.JSON(409, ErrUserExists)
			return
		}
		c.JSON(500, ErrDatabase)
		return
	}
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	c.JSON(200, NewRegisterResponse(user.ID, user.Username, email))
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}
	user, err := h.userService.GetUserByUsername(c, req.Username)
	if err != nil {
		c.JSON(401, ErrInvalidCredentials)
		return
	}

	if !util.VerifyPassword(req.Password, user.PasswordHash) {
		c.JSON(401, ErrInvalidCredentials)
		return
	}
	// 生成 JWT
	token, err := util.GenerateToken(user.ID, user.Username, user.Role, 24*time.Hour)
	if err != nil {
		c.JSON(500, ErrGenerateJWT)
		return
	}

	c.JSON(200, NewLoginResponse(token, 24*3600))
}
