package admin

// BaseResponse 基础响应结构
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// UserResponse 用户操作响应
type UserResponse struct {
	BaseResponse
	UserID    string `json:"user_id,omitempty"`
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// LinkResponse 链接操作响应
type LinkResponse struct {
	BaseResponse
	LinkID      string `json:"link_id,omitempty"`
	ShortCode   string `json:"short_code,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	BaseResponse
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// ListLinksResponse 链接列表响应
type ListLinksResponse struct {
	BaseResponse
	Links []LinkResponse `json:"links"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
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
func NewCreateUserResponse(userID, username, email string) UserResponse {
	return UserResponse{
		BaseResponse: NewSuccessResponse("用户创建成功"),
		UserID:       userID,
		Username:     username,
		Email:        email,
	}
}

func NewDeleteUserResponse(userID string) UserResponse {
	return UserResponse{
		BaseResponse: NewSuccessResponse("用户删除成功"),
		UserID:       userID,
	}
}

func NewActivateUserResponse(userID string, isActive bool) UserResponse {
	return UserResponse{
		BaseResponse: NewSuccessResponse("用户状态更新成功"),
		UserID:       userID,
		IsActive:     &isActive,
	}
}

func NewUnactivateUserResponse(userID string, isActive bool) UserResponse {
	return NewActivateUserResponse(userID, isActive) // 复用同一个函数
}

// 链接相关响应构造函数
func NewActivateLinkResponse(linkID string, isActive bool) LinkResponse {
	return LinkResponse{
		BaseResponse: NewSuccessResponse("链接状态更新成功"),
		LinkID:       linkID,
		IsActive:     &isActive,
	}
}

func NewUnactivateLinkResponse(linkID string, isActive bool) LinkResponse {
	return NewActivateLinkResponse(linkID, isActive) // 复用同一个函数
}

// 列表响应构造函数
func NewListUsersResponse(users []UserResponse, total, page, limit int) ListUsersResponse {
	return ListUsersResponse{
		BaseResponse: NewSuccessResponse("获取用户列表成功"),
		Users:        users,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}
}

func NewListLinksResponse(links []LinkResponse, total, page, limit int) ListLinksResponse {
	return ListLinksResponse{
		BaseResponse: NewSuccessResponse("获取链接列表成功"),
		Links:        links,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}
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
	ErrInvalidRequest    = NewErrorResponse("BAD_REQUEST", "请求无效", "")
	ErrUserNotFound      = NewErrorResponse("USER_NOT_FOUND", "用户不存在", "")
	ErrLinkNotFound      = NewErrorResponse("LINK_NOT_FOUND", "链接不存在", "")
	ErrUserAlreadyExists = NewErrorResponse("USER_EXISTS", "用户已存在", "")
	ErrForbidden         = NewErrorResponse("FORBIDDEN", "没有操作权限", "")
	ErrUnauthorized      = NewErrorResponse("UNAUTHORIZED", "未授权访问", "")

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
