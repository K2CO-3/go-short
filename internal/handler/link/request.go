package link

import "time"

// CreateLinkRequest 创建短链接请求
type CreateLinkRequest struct {
	URL       string     `json:"url" binding:"required,url"`
	Alias     *string    `json:"alias"`
	ExpiresAt *time.Time `json:"expires_at" binding:"omitempty,expiry_time"`
	Status    *bool      `json:"status"`
	ShortCode *string    `json:"short_code" binding:"omitempty,short_code"`
}
