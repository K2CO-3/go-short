package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Username     string    `gorm:"not null;unique;size:64"`
	PasswordHash string    `gorm:"not null;column:password_hash"`
	Email        *string   `gorm:"size:128"`
	Role         string    `gorm:"size:20;default:'user'"`
	Status       string    `gorm:"size:20;default:'active'"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
	LinkCount    int64     `gorm:"default:0"`
}

func (User) TableName() string {
	return "users"
}
