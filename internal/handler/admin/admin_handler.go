package admin

import (
	"errors"
	"go-short/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminService *service.AdminService // 通过依赖注入，不用自己连接
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}
	createUserRequest := service.CreateUserCommand{
		Username: req.Username,
		Password: req.Password,
		Status:   req.Status,
		Email:    req.Email,
		Role:     req.Role,
	}
	user, err := h.adminService.CreateUser(c, createUserRequest)
	if user == nil || err != nil {
		if errors.Is(err, service.ErrUserExists) {
			c.JSON(409, ErrUserAlreadyExists)
			return
		}
		c.JSON(500, ErrDatabase)
		return
	}
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	c.JSON(200, NewCreateUserResponse(user.ID.String(), user.Username, email))
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userIDstr := c.Param("userID")
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}

	// 使用 AdminService 的事务方法删除用户及其所有数据
	if err := h.adminService.DeleteUserAndData(c, userID); err != nil {
		c.JSON(500, ErrDatabase)
		return
	}
	c.JSON(200, NewDeleteUserResponse(userIDstr))
}

func (h *AdminHandler) UnactiveUser(c *gin.Context) {
	userIDstr := c.Param("userID")
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}
	if err := h.adminService.UnactiveUserByUserID(c, userID); err != nil {
		c.JSON(500, ErrDatabase)
		return
	}
	c.JSON(200, NewUnactivateUserResponse(userIDstr, false))
}

func (h *AdminHandler) ActiveUser(c *gin.Context) {
	userIDstr := c.Param("userID")
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}
	if err := h.adminService.ActiveUserByUserID(c, userID); err != nil {
		c.JSON(500, ErrDatabase)
		return
	}
	c.JSON(200, NewActivateUserResponse(userIDstr, true))
}

// GetUsers 获取所有用户列表（分页）
func (h *AdminHandler) GetUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	users, total, err := h.adminService.GetAllUsers(c, page, size)
	if err != nil {
		c.JSON(500, ErrDatabase)
		return
	}

	// 转换为响应格式
	userResponses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		email := ""
		if user.Email != nil {
			email = *user.Email
		}

		isActive := user.Status == "active"
		createdAt := ""
		updatedAt := ""
		if !user.CreatedAt.IsZero() {
			createdAt = user.CreatedAt.Format("2006-01-02T15:04:05Z")
		}
		if !user.UpdatedAt.IsZero() {
			updatedAt = user.UpdatedAt.Format("2006-01-02T15:04:05Z")
		}

		userResponses = append(userResponses, UserResponse{
			UserID:    user.ID.String(),
			Username:  user.Username,
			Email:     email,
			IsActive:  &isActive,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	c.JSON(200, NewListUsersResponse(userResponses, int(total), page, size))
}

func (h *AdminHandler) ActiveLink(c *gin.Context) {
	linkIDStr := c.Param("linkID")
	linkID, err := strconv.ParseInt(linkIDStr, 10, 64)
	if err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}
	if err := h.adminService.ActiveLink(c, linkID); err != nil {
		c.JSON(500, ErrDatabase)
		return
	}
	c.JSON(200, NewActivateLinkResponse(linkIDStr, true))
}

func (h *AdminHandler) UnactiveLink(c *gin.Context) {
	linkIDStr := c.Param("linkID")
	linkID, err := strconv.ParseInt(linkIDStr, 10, 64)
	if err != nil {
		c.JSON(400, ErrInvalidRequest)
		return
	}
	if err := h.adminService.UnactiveLink(c, linkID); err != nil {
		c.JSON(500, ErrDatabase)
		return
	}
	c.JSON(200, NewUnactivateLinkResponse(linkIDStr, false))
}
