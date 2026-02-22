package link

import (
	"go-short/internal/model"
	"time"
)

// BaseResponse 基础响应结构
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// LinkResponse 链接操作响应
type LinkResponse struct {
	BaseResponse
	LinkID      int64      `json:"link_id,omitempty"`
	ShortCode   string     `json:"short_code,omitempty"`
	OriginalURL string     `json:"original_url,omitempty"`
	ShortURL    string     `json:"short_url,omitempty"`
	Alias       string     `json:"alias,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
}

// ListLinksResponse 链接列表响应
type ListLinksResponse struct {
	BaseResponse
	Links []LinkResponse `json:"links"`
	Total int64          `json:"total"`
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

// 链接相关响应构造函数
func NewCreateLinkResponse(link *model.Link, shortURL string) LinkResponse {
	isActive := link.Status
	return LinkResponse{
		BaseResponse: NewSuccessResponse("短链接创建成功"),
		LinkID:       link.ID,
		ShortCode:    link.ShortCode,
		OriginalURL:  link.OriginalURL,
		ShortURL:     shortURL,
		Alias:        link.Alias,
		IsActive:     &isActive,
		ExpiresAt:    link.ExpiresAt,
		CreatedAt:    link.CreatedAt,
	}
}

func NewLinkExistsResponse(link *model.Link, shortURL string) LinkResponse {
	isActive := link.Status
	return LinkResponse{
		BaseResponse: NewSuccessResponse("链接已存在"),
		LinkID:       link.ID,
		ShortCode:    link.ShortCode,
		OriginalURL:  link.OriginalURL,
		ShortURL:     shortURL,
		Alias:        link.Alias,
		IsActive:     &isActive,
		ExpiresAt:    link.ExpiresAt,
		CreatedAt:    link.CreatedAt,
	}
}

func NewListLinksResponse(links []model.Link, total int64, page, limit int, baseURL string) ListLinksResponse {
	linkResponses := make([]LinkResponse, 0, len(links))
	for _, link := range links {
		isActive := link.Status
		shortURL := ""
		if link.ShortCode != "" {
			shortURL = baseURL + "/code/" + link.ShortCode
		}
		linkResponses = append(linkResponses, LinkResponse{
			LinkID:      link.ID,
			ShortCode:   link.ShortCode,
			OriginalURL: link.OriginalURL,
			ShortURL:    shortURL,
			Alias:       link.Alias,
			IsActive:    &isActive,
			ExpiresAt:   link.ExpiresAt,
			CreatedAt:   link.CreatedAt,
		})
	}
	return ListLinksResponse{
		BaseResponse: NewSuccessResponse("获取链接列表成功"),
		Links:        linkResponses,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}
}

func NewDeleteLinkResponse() BaseResponse {
	return NewSuccessResponse("短链接删除成功")
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
	ErrInvalidRequest     = NewErrorResponse("BAD_REQUEST", "请求无效", "")
	ErrInvalidUserID      = NewErrorResponse("INVALID_USER_ID", "无效的用户ID", "")
	ErrLinkNotFound       = NewErrorResponse("LINK_NOT_FOUND", "链接不存在", "")
	ErrLinkAlreadyExists  = NewErrorResponse("LINK_EXISTS", "链接已存在", "")
	ErrInvalidExpiryTime  = NewErrorResponse("INVALID_EXPIRY_TIME", "过期时间无效", "")
	ErrShortCodeDuplicate = NewErrorResponse("SHORT_CODE_DUPLICATE", "短码已被占用", "")
	ErrForbidden          = NewErrorResponse("FORBIDDEN", "没有操作权限", "")
	ErrUnauthorized       = NewErrorResponse("UNAUTHORIZED", "未授权访问", "")
	ErrUserNotFound       = NewErrorResponse("USER_NOT_FOUND", "用户不存在", "")
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
