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
)

var (
	ErrLinkAlreadyExists = errors.New("链接已存在")
	ErrShortCodeExists   = errors.New("短码已被占用")
	ErrLinkNotFound      = errors.New("链接不存在")
	ErrForbidden         = errors.New("没有操作权限")
	ErrUserNotFound      = errors.New("用户不存在")
)

type CreateLinkCommand struct {
	OriginalURL string
	Alias       *string
	ShortCode   *string
	Status      *bool
	ExpiresAt   *time.Time
	UserID      uuid.UUID
}

// CreateLink 创建短链接（包含所有业务逻辑）
func (s *LinkService) CreateLink(ctx context.Context, cmd CreateLinkCommand) (*model.Link, error) {
	// 1. 规范化 URL
	normalizedURL := strings.TrimSpace(cmd.OriginalURL)
	if normalizedURL == "" {
		return nil, fmt.Errorf("URL不能为空")
	}
	if !strings.HasPrefix(normalizedURL, "http://") && !strings.HasPrefix(normalizedURL, "https://") {
		normalizedURL = "https://" + normalizedURL
	}

	// 2. 检查自定义短码是否已被占用
	if cmd.ShortCode != nil && *cmd.ShortCode != "" {
		exists, err := s.linkRepository.CheckShortCodeDuplicate(ctx, s.db, *cmd.ShortCode)
		if err != nil {
			return nil, fmt.Errorf("检查短码失败: %w", err)
		}
		if exists {
			return nil, ErrShortCodeExists
		}
	}

	// 3. 检查用户是否已经创建过相同的链接
	existingLink, err := s.linkRepository.GetLinkByUserAndURL(ctx, s.db, cmd.UserID, normalizedURL)
	if err == nil && existingLink != nil {
		// 返回已存在的链接和特殊错误
		return existingLink, ErrLinkAlreadyExists
	}

	// 4. 处理别名
	alias := ""
	if cmd.Alias != nil && *cmd.Alias != "" {
		alias = *cmd.Alias
	} else {
		// 查询用户已有链接数，生成默认别名
		count, err := s.linkRepository.GetNumOfLinksByUser(ctx, s.db, cmd.UserID)
		if err != nil {
			return nil, fmt.Errorf("获取链接数失败: %w", err)
		}
		alias = fmt.Sprintf("短链接%d", count+1)
	}

	// 5. 创建链接对象
	link := &model.Link{
		OriginalURL: normalizedURL,
		UserID:      cmd.UserID,
		CreatedAt:   time.Now(),
		ExpiresAt:   cmd.ExpiresAt,
		Status:      *cmd.Status,
		Alias:       alias,
	}

	if cmd.Status != nil {
		link.Status = *cmd.Status
	}

	// 6. 处理自定义短码
	if cmd.ShortCode != nil && *cmd.ShortCode != "" {
		link.ShortCode = *cmd.ShortCode
		link.IsCustom = true
	}

	// 7. 存入数据库
	if err := s.linkRepository.Create(ctx, s.db, link); err != nil {
		return nil, fmt.Errorf("创建链接失败: %w", err)
	}

	// 8. 如果短码未设置，使用 ID 生成短码并更新
	if link.ShortCode == "" && link.ID > 0 {
		link.ShortCode = util.Encode(link.ID)
		if err := s.linkRepository.Update(ctx, s.db, link); err != nil {
			return nil, fmt.Errorf("更新短码失败: %w", err)
		}
	}

	return link, nil
}

func (s *LinkService) Create(ctx context.Context, link *model.Link) error {
	return s.linkRepository.Create(ctx, s.db, link)
}

func (s *LinkService) Update(ctx context.Context, link *model.Link) error {
	return s.linkRepository.Update(ctx, s.db, link)
}

func (s *LinkService) GetLinkByCode(ctx context.Context, code string) (*model.Link, error) {
	return s.linkRepository.GetLinkByCode(ctx, s.db, code)
}

// GetLinkByCodeForRedirect 根据短码获取链接（用于重定向服务，包含过期时间检查）
func (s *LinkService) GetLinkByCodeForRedirect(ctx context.Context, code string) (*model.Link, error) {
	link, err := s.linkRepository.GetLinkByCode(ctx, s.db, code)
	if err != nil {
		return nil, ErrLinkNotFound
	}

	// 检查链接是否过期
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		return nil, ErrLinkNotFound // 过期视为不存在
	}

	return link, nil
}

func (s *LinkService) GetLinkByUserAndURL(ctx context.Context, userID uuid.UUID, originalURL string) (*model.Link, error) {
	return s.linkRepository.GetLinkByUserAndURL(ctx, s.db, userID, originalURL)
}

func (s *LinkService) GetLinksByUser(ctx context.Context, userID uuid.UUID, page, size int) ([]model.Link, int64, error) {
	return s.linkRepository.GetLinksByUser(ctx, s.db, userID, page, size)
}

func (s *LinkService) GetNumOfLinksByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.linkRepository.GetNumOfLinksByUser(ctx, s.db, userID)
}

func (s *LinkService) CheckShortCodeDuplicate(ctx context.Context, short_code string) (bool, error) {
	return s.linkRepository.CheckShortCodeDuplicate(ctx, s.db, short_code)
}

func (s *LinkService) ActiveLink(ctx context.Context, linkID int64) error {
	return s.linkRepository.ActiveLink(ctx, s.db, linkID)
}

func (s *LinkService) UnactiveLink(ctx context.Context, linkID int64) error {
	return s.linkRepository.UnactiveLink(ctx, s.db, linkID)
}

func (s *LinkService) GetLinksByUserAlias(ctx context.Context, userID uuid.UUID, alias string, page, size int) ([]model.Link, int64, error) {
	return s.linkRepository.GetLinksByUserAlias(ctx, s.db, userID, alias, page, size)
}

// DeleteLink 根据ID删除短链接（需要验证用户权限）
func (s *LinkService) DeleteLink(ctx context.Context, linkID int64, userID uuid.UUID, isAdmin bool) error {
	link, err := s.linkRepository.GetLinkByID(ctx, s.db, linkID)
	if err != nil {
		return ErrLinkNotFound
	}

	if link.UserID != userID && !isAdmin {
		return ErrForbidden
	}

	if err := s.linkRepository.DeleteLinkByID(ctx, s.db, linkID); err != nil {
		return fmt.Errorf("删除链接失败: %w", err)
	}

	return nil
}
