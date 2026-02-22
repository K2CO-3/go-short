package auth

import "github.com/google/uuid"

// BaseResponse 基础响应结构
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	BaseResponse
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	BaseResponse
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type ErrorResponse struct {
	BaseResponse
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func NewSuccessResponse(message string) BaseResponse {
	return BaseResponse{
		Success: true,
		Message: message,
	}
}

func NewRegisterResponse(userID uuid.UUID, username, email string) RegisterResponse {
	resp := RegisterResponse{
		BaseResponse: NewSuccessResponse("User registered successfully"),
	}
	resp.UserID = userID
	resp.Username = username
	resp.Email = email
	return resp
}

func NewLoginResponse(accessToken string, expiresIn int) LoginResponse {
	resp := LoginResponse{
		BaseResponse: NewSuccessResponse("Login successful"),
	}
	resp.AccessToken = accessToken
	resp.ExpiresIn = expiresIn
	resp.TokenType = "Bearer"
	return resp
}

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

type ErrorCode string

var (
	// 客户端错误 (4xx) - 业务逻辑错误
	ErrInvalidRequest     = NewErrorResponse("BAD_REQUEST", "请求无效", "")
	ErrUserExists         = NewErrorResponse("USER_EXISTS", "用户已存在", "")
	ErrInvalidCredentials = NewErrorResponse("INVALID_CREDENTIALS", "用户名或密码错误", "")

	// 服务器错误 (5xx) - 系统错误
	ErrPasswordHash = NewErrorResponse("INTERNAL_ERROR", "系统错误", "")
	ErrDatabase     = NewErrorResponse("INTERNAL_ERROR", "系统错误", "")
	ErrInternal     = NewErrorResponse("INTERNAL_ERROR", "系统错误", "")
	ErrGenerateJWT  = NewErrorResponse("INTERNAL_ERROR", "系统错误", "")
)

func NewError(code ErrorCode, message string) ErrorResponse {
	return ErrorResponse{
		BaseResponse: BaseResponse{Success: false},
		Code:         string(code),
		Message:      message,
	}
}
