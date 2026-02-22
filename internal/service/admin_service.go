package service

import (
	"context"
	"fmt"
	"go-short/internal/model"
	"go-short/internal/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateUserCommand struct {
	Username string
	Email    string
	Password string
	Role     string
	Status   string
}

func (s *AdminService) CreateUser(ctx context.Context, cmd CreateUserCommand) (*model.User, error) {
	// 1. 检查用户名是否已存在
	exists, err := s.userRepository.CheckUsernameExists(ctx, s.db, cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	// 2. 加密密码
	hashedPassword, err := util.HashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 3. 生成用户ID
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("生成用户ID失败: %w", err)
	}

	// 4. 创建用户对象
	user := model.User{
		ID:           id,
		Username:     cmd.Username,
		Email:        &cmd.Email,
		Role:         cmd.Role,
		PasswordHash: hashedPassword,
		Status:       cmd.Status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LinkCount:    0,
	}

	// 5. 保存到数据库
	if err := s.userRepository.Create(ctx, s.db, &user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return &user, nil
}

// DeleteUser 删除用户（仅删除用户记录）
func (s *AdminService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepository.DeleteUserByID(ctx, s.db, userID)
}

// DeleteUserAndData 删除用户及其所有关联数据（使用事务保证原子性）
func (s *AdminService) DeleteUserAndData(ctx context.Context, userID uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除用户的所有链接
		if err := s.linkRepository.DeleteLinksByUser(ctx, tx, userID); err != nil {
			return err
		}

		// 2. 删除用户记录
		if err := s.userRepository.DeleteUserByID(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

// UnactiveUserByUserID 禁用用户
func (s *AdminService) UnactiveUserByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.userRepository.UnactiveUserByUserID(ctx, s.db, userID)
}

// ActiveUserByUserID 激活用户
func (s *AdminService) ActiveUserByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.userRepository.ActiveUserByUserID(ctx, s.db, userID)
}

// GetAllUsers 获取所有用户列表（分页）
func (s *AdminService) GetAllUsers(ctx context.Context, page, size int) ([]model.User, int64, error) {
	return s.userRepository.GetAllUsers(ctx, s.db, page, size)
}

// ActiveLink 激活链接
func (s *AdminService) ActiveLink(ctx context.Context, linkID int64) error {
	return s.linkRepository.ActiveLink(ctx, s.db, linkID)
}

// UnactiveLink 禁用链接
func (s *AdminService) UnactiveLink(ctx context.Context, linkID int64) error {
	return s.linkRepository.UnactiveLink(ctx, s.db, linkID)
}
