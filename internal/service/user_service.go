package service

import (
	"context"
	"errors"
	"fmt"
	"go-short/internal/model"
	"go-short/internal/util"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidPassword = errors.New("旧密码错误")
	ErrUserExists      = errors.New("用户名已存在")
)

type RegisterUserCommand struct {
	Username string
	Email    string
	Password string
}

func (s *UserService) CreateUser(ctx context.Context, cmd RegisterUserCommand) (*model.User, error) {
	hashedPassword, err := util.HashPassword(cmd.Password)
	if err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	user := model.User{
		ID:           id,
		Username:     cmd.Username,
		PasswordHash: string(hashedPassword),
		Email:        &cmd.Email,
		Role:         "user",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LinkCount:    0,
	}
	if err := s.userRepository.Create(ctx, s.db, &user); err != nil {
		// 检查是否是唯一约束违反错误（用户名重复）
		if errors.Is(err, gorm.ErrDuplicatedKey) || 
		   strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
		   strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrUserExists
		}
		return nil, err
	}
	return &user, nil
}
func (s *UserService) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepository.GetUserByUserID(ctx, s.db, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.userRepository.GetUserByUsername(ctx, s.db, username)
}

func (s *UserService) UnactiveUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepository.UnactiveUserByUserID(ctx, s.db, userID)
}

func (s *UserService) ActiveUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepository.ActiveUserByUserID(ctx, s.db, userID)
}

type UpdateUserProfileCommand struct {
	Username string
	Email    *string
}

// UpdateUserProfile 更新用户个人信息（通用接口）
func (s *UserService) UpdateUserProfile(ctx context.Context, cmd UpdateUserProfileCommand) (*model.User, error) {
	// 先获取用户信息
	user, err := s.userRepository.GetUserByUsername(ctx, s.db, cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 更新邮箱（如果提供）
	if cmd.Email != nil {
		user.Email = cmd.Email
	}

	// 保存更新
	if err := s.userRepository.Update(ctx, s.db, user); err != nil {
		return nil, fmt.Errorf("更新用户信息失败: %w", err)
	}

	return user, nil
}

// UpdatePassword 修改密码
func (s *UserService) UpdatePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	// 先获取用户信息
	user, err := s.userRepository.GetUserByUsername(ctx, s.db, username)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 验证旧密码
	if !util.VerifyPassword(oldPassword, user.PasswordHash) {
		return ErrInvalidPassword
	}

	// 加密新密码
	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepository.Update(ctx, s.db, user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}
