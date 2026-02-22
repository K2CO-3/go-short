package user

import (
	"errors"
	"go-short/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile 获取当前用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, ErrUnauthorized)
		return
	}

	usernameStr, ok := username.(string)
	if !ok {
		c.JSON(401, ErrUnauthorized)
		return
	}

	user, err := h.userService.GetUserByUsername(c, usernameStr)
	if err != nil {
		c.JSON(500, ErrDatabase)
		return
	}

	c.JSON(200, NewUserResponse(user))
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, ErrUnauthorized)
		return
	}

	usernameStr, ok := username.(string)
	if !ok {
		c.JSON(401, ErrUnauthorized)
		return
	}

	cmd := service.UpdateUserProfileCommand{
		Username: usernameStr,
		Email:    req.Email,
	}

	user, err := h.userService.UpdateUserProfile(c, cmd)
	if err != nil {
		c.JSON(500, ErrDatabase)
		return
	}

	c.JSON(200, NewUpdateUserResponse(user))
}

// UpdatePassword 修改密码
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, ErrUnauthorized)
		return
	}

	usernameStr, ok := username.(string)
	if !ok {
		c.JSON(401, ErrUnauthorized)
		return
	}

	err := h.userService.UpdatePassword(c, usernameStr, req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(400, ErrInvalidPassword)
			return
		}
		c.JSON(500, ErrDatabase)
		return
	}

	c.JSON(200, NewUpdatePasswordResponse())
}
