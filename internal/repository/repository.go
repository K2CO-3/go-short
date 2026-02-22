package repository

import (
	"context"
	"go-short/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, tx *gorm.DB, user *model.User) error
	Update(ctx context.Context, tx *gorm.DB, user *model.User) error
	DeleteUserByID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error
	GetUserByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) (*model.User, error)
	GetUserByUsername(ctx context.Context, tx *gorm.DB, username string) (*model.User, error)
	CheckUsernameExists(ctx context.Context, tx *gorm.DB, username string) (bool, error)
	GetAllUsers(ctx context.Context, tx *gorm.DB, page, size int) ([]model.User, int64, error)
	UnactiveUserByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error
	ActiveUserByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error
	UpdatePasswordByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID, password string) error
}

type LinkRepository interface {
	Create(ctx context.Context, tx *gorm.DB, link *model.Link) error
	Update(ctx context.Context, tx *gorm.DB, link *model.Link) error
	GetLinkByCode(ctx context.Context, tx *gorm.DB, code string) (*model.Link, error)
	CheckShortCodeExists(ctx context.Context, tx *gorm.DB, code string) (bool, error)
	GetLinkByUserAndURL(ctx context.Context, tx *gorm.DB, userID uuid.UUID, originalURL string) (*model.Link, error)
	GetLinksByUser(ctx context.Context, tx *gorm.DB, userID uuid.UUID, page, size int) ([]model.Link, int64, error)
	GetLinksByUserAlias(ctx context.Context, tx *gorm.DB, userID uuid.UUID, alias string, page, size int) ([]model.Link, int64, error)
	GetLinkIDByCode(ctx context.Context, tx *gorm.DB, code string) (int64, error)
	ActiveLink(ctx context.Context, tx *gorm.DB, LinkID int64) error
	UnactiveLink(ctx context.Context, tx *gorm.DB, LinkID int64) error
	GetNumOfLinksByUser(ctx context.Context, tx *gorm.DB, userID uuid.UUID) (int64, error)
	CheckShortCodeDuplicate(ctx context.Context, tx *gorm.DB, short_code string) (bool, error)
	DeleteLinksByUser(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error
	GetLinkByID(ctx context.Context, tx *gorm.DB, linkID int64) (*model.Link, error)
	DeleteLinkByID(ctx context.Context, tx *gorm.DB, linkID int64) error
}

type AccessLogRepository interface {
	SaveAccessLog(ctx context.Context, tx *gorm.DB, logEntry *model.AccessLog) error
}
