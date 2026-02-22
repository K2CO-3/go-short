package user

import (
	"go-short/internal/model"
	"time"

	"github.com/google/uuid"
)

// BaseResponse 基础响应结构
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	BaseResponse
	UserID    uuid.UUID  `json:"user_id,omitempty"`
	Username  string     `json:"username,omitempty"`
	Email     *string    `json:"email,omitempty"`
	Role      string     `json:"role,omitempty"`
	Status    string     `json:"status,omitempty"`
	LinkCount int64      `json:"link_count,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	BaseResponse
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// 成功响应构造函数
func NewSuccessResponse(message string) BaseResponse {
	return BaseResponse{
		Success: true,
		Message: message,
	}
}

// 用户相关响应构造函数
func NewUserResponse(user *model.User) UserResponse {
	// 检查时间字段是否为零值，如果是则设为 nil（避免返回零值时间）
	var createdAt *time.Time
	var updatedAt *time.Time

	if !user.CreatedAt.IsZero() {
		createdAt = &user.CreatedAt
	}
	if !user.UpdatedAt.IsZero() {
		updatedAt = &user.UpdatedAt
	}

	return UserResponse{
		BaseResponse: NewSuccessResponse("获取用户信息成功"),
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role,
		Status:       user.Status,
		LinkCount:    user.LinkCount,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func NewUpdateUserResponse(user *model.User) UserResponse {
	// 检查时间字段是否为零值，如果是则设为 nil（避免返回零值时间）
	var createdAt *time.Time
	var updatedAt *time.Time

	if !user.CreatedAt.IsZero() {
		createdAt = &user.CreatedAt
	}
	if !user.UpdatedAt.IsZero() {
		updatedAt = &user.UpdatedAt
	}

	return UserResponse{
		BaseResponse: NewSuccessResponse("用户信息更新成功"),
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role,
		Status:       user.Status,
		LinkCount:    user.LinkCount,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func NewUpdatePasswordResponse() BaseResponse {
	return NewSuccessResponse("密码修改成功")
}

// 错误响应构造函数
func NewErrorResponse(code, message, details string) ErrorResponse {
	return ErrorResponse{
		BaseResponse: BaseResponse{
			Success: false,
		},
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 错误码定义
type ErrorCode string

var (
	// 客户端错误 (4xx) - 业务逻辑错误
	ErrInvalidRequest   = NewErrorResponse("BAD_REQUEST", "请求无效", "")
	ErrUserNotFound     = NewErrorResponse("USER_NOT_FOUND", "用户不存在", "")
	ErrInvalidPassword  = NewErrorResponse("INVALID_PASSWORD", "密码错误", "")
	ErrPasswordTooShort = NewErrorResponse("PASSWORD_TOO_SHORT", "密码长度不足", "")
	ErrForbidden        = NewErrorResponse("FORBIDDEN", "没有操作权限", "")
	ErrUnauthorized     = NewErrorResponse("UNAUTHORIZED", "未授权访问", "")

	// 服务器错误 (5xx) - 系统错误
	ErrDatabase = NewErrorResponse("INTERNAL_ERROR", "系统错误", "")
	ErrInternal = NewErrorResponse("INTERNAL_ERROR", "系统错误", "")
)

func NewError(code ErrorCode, message string) ErrorResponse {
	return ErrorResponse{
		BaseResponse: BaseResponse{Success: false},
		Code:         string(code),
		Message:      message,
	}
}
