package model

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID          int64      `gorm:"primaryKey"`
	ShortCode   string     `gorm:"not null;unique;size:20;default:''"`
	OriginalURL string     `gorm:"not null;type:text"`
	Alias       string     `gorm:"size:100;default:''"`
	UserID      uuid.UUID  `gorm:"index:idx_links_user_id"`
	IsCustom    bool       `gorm:"default:false"`
	VisitCount  int64      `gorm:"default:0"`
	ExpiresAt   *time.Time `gorm:"index:idx_links_expires_at"`
	Status      bool       `gorm:"default:true"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (Link) TableName() string {
	return "links"
}
