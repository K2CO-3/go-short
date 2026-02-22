package postgresql

import (
	"context"
	"go-short/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepoImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 实例
func NewUserRepository(db *gorm.DB) *userRepoImpl {
	return &userRepoImpl{db: db}
}

// ==========================================
// User 相关操作 (API Service)
// ==========================================

// Create 创建用户
func (d *userRepoImpl) Create(ctx context.Context, tx *gorm.DB, user *model.User) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Create(user).Error
}

// Update 更新用户信息
func (d *userRepoImpl) Update(ctx context.Context, tx *gorm.DB, user *model.User) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (d *userRepoImpl) DeleteUserByID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Where("id = ?", userID).Delete(&model.User{}).Error
}

// GetUserByUserID 根据用户ID查找用户
func (d *userRepoImpl) GetUserByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) (*model.User, error) {
	if tx == nil {
		tx = d.db
	}
	var user model.User
	err := tx.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名查找用户 (用于登录)
func (d *userRepoImpl) GetUserByUsername(ctx context.Context, tx *gorm.DB, username string) (*model.User, error) {
	if tx == nil {
		tx = d.db
	}
	var user model.User
	err := tx.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		// 直接返回错误，让调用方处理 gorm.ErrRecordNotFound
		return nil, err
	}
	return &user, nil
}

// CheckUsernameExists 检查用户名是否已存在
func (d *userRepoImpl) CheckUsernameExists(ctx context.Context, tx *gorm.DB, username string) (bool, error) {
	if tx == nil {
		tx = d.db
	}
	var count int64
	err := tx.WithContext(ctx).Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAllUsers 获取所有用户列表（分页）
func (d *userRepoImpl) GetAllUsers(ctx context.Context, tx *gorm.DB, page, size int) ([]model.User, int64, error) {
	if tx == nil {
		tx = d.db
	}
	var users []model.User
	var total int64
	offset := (page - 1) * size

	// 统计总数 - 使用独立的查询对象
	baseQuery := tx.WithContext(ctx).Model(&model.User{})
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表 - 创建新的查询对象，避免复用导致的状态问题
	if err := tx.WithContext(ctx).
		Model(&model.User{}).
		Offset(offset).
		Limit(size).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UnactiveUserByUserID 禁用用户
func (d *userRepoImpl) UnactiveUserByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("status", "unactivate").Error
}

// ActiveUserByUserID 激活用户
func (d *userRepoImpl) ActiveUserByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("status", "active").Error
}
func (d *userRepoImpl) UpdatePasswordByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID, password string) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("password_hash", password).Error
}
