package model

import "time"

type AccessLog struct {
	ID        int64     `gorm:"primaryKey"`
	LinkID    int64     `gorm:"index:idx_access_logs_link_id"`
	ShortCode string    `gorm:"not null;size:20"`
	IPAddress string    `gorm:"size:45"` // 支持IPv6
	UserAgent string    `gorm:"type:text"`
	Referer   string    `gorm:"type:text"`
	VisitedAt time.Time `gorm:"autoCreateTime;index:,sort:desc"`
}

// TableName 指定表名
func (AccessLog) TableName() string {
	return "access_logs"
}

// 创建复合索引
func (AccessLog) Indexes() []map[string][]string {
	return []map[string][]string{
		{
			"idx_logs_code_time": {"ShortCode", "VisitedAt:desc"},
		},
	}
}
