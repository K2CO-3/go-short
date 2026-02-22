package postgresql

import (
	"context"
	"go-short/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type linkRepoImpl struct {
	db *gorm.DB
}

// NewLinkRepository 创建 LinkRepository 实例
func NewLinkRepository(db *gorm.DB) *linkRepoImpl {
	return &linkRepoImpl{db: db}
}

// ==========================================
// Link 相关操作
// ==========================================

// Create 创建短链接记录
func (d *linkRepoImpl) Create(ctx context.Context, tx *gorm.DB, link *model.Link) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Create(link).Error
}

// Update 更新短链接记录
func (d *linkRepoImpl) Update(ctx context.Context, tx *gorm.DB, link *model.Link) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Save(link).Error
}

// GetLinkByCode 根据短码查询链接 (用于重定向服务，通常这里会有 DB 级的 Fallback)
func (d *linkRepoImpl) GetLinkByCode(ctx context.Context, tx *gorm.DB, code string) (*model.Link, error) {
	if tx == nil {
		tx = d.db
	}
	var link model.Link
	// 查询条件：短码匹配且状态为启用
	err := tx.WithContext(ctx).
		Where("short_code = ? AND status = ?", code, true).
		First(&link).Error

	if err != nil {
		return nil, err
	}
	return &link, nil
}

// CheckShortCodeExists 检查短码是否已被占用 (用于自定义短码)
func (d *linkRepoImpl) CheckShortCodeExists(ctx context.Context, tx *gorm.DB, code string) (bool, error) {
	if tx == nil {
		tx = d.db
	}
	var count int64
	err := tx.WithContext(ctx).Model(&model.Link{}).
		Where("short_code = ?", code).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetLinkByUserAndURL 根据用户ID和原始URL查询链接（用于检查是否已存在）
func (d *linkRepoImpl) GetLinkByUserAndURL(ctx context.Context, tx *gorm.DB, userID uuid.UUID, originalURL string) (*model.Link, error) {
	if tx == nil {
		tx = d.db
	}
	var link model.Link
	err := tx.WithContext(ctx).
		Where("user_id = ? AND original_url = ? AND status = ?", userID, originalURL, true).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// GetLinksByUser 获取用户的链接列表 (分页)
func (d *linkRepoImpl) GetLinksByUser(ctx context.Context, tx *gorm.DB, userID uuid.UUID, page, size int) ([]model.Link, int64, error) {
	if tx == nil {
		tx = d.db
	}
	var links []model.Link
	var total int64
	offset := (page - 1) * size

	// 统计总数
	query := tx.WithContext(ctx).Model(&model.Link{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&links).Error; err != nil {
		return nil, 0, err
	}

	return links, total, nil
}

// GetLinkIDByCode 根据短码查询链接ID (用于日志查询服务)
func (d *linkRepoImpl) GetLinkIDByCode(ctx context.Context, tx *gorm.DB, code string) (int64, error) {
	if tx == nil {
		tx = d.db
	}
	var linkID int64
	// 查询条件：短码匹配且状态为启用
	err := tx.WithContext(ctx).
		Model(&model.Link{}).
		Where("short_code = ? AND status = ?", code, true).
		Select("id").
		First(&linkID).Error

	if err != nil {
		return 0, err
	}
	return linkID, nil
}

func (d *linkRepoImpl) GetLinksByUserAlias(ctx context.Context, tx *gorm.DB, userID uuid.UUID, alias string, page, size int) ([]model.Link, int64, error) {
	if tx == nil {
		tx = d.db
	}
	var links []model.Link
	var total int64
	offset := (page - 1) * size

	query := tx.WithContext(ctx).Model(&model.Link{}).Where("alias = ? AND user_id = ?", alias, userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&links).Error; err != nil {
		return nil, 0, err
	}
	return links, total, nil
}

func (d *linkRepoImpl) DeleteLinksByUser(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Link{}).Error
}

func (d *linkRepoImpl) ActiveLink(ctx context.Context, tx *gorm.DB, LinkID int64) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Model(&model.Link{}).Where("id = ?", LinkID).Update("status", true).Error
}

func (d *linkRepoImpl) UnactiveLink(ctx context.Context, tx *gorm.DB, LinkID int64) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Model(&model.Link{}).Where("id = ?", LinkID).Update("status", false).Error
}
func (d *linkRepoImpl) GetNumOfLinksByUser(ctx context.Context, tx *gorm.DB, userID uuid.UUID) (int64, error) {
	if tx == nil {
		tx = d.db
	}
	var count int64
	err := tx.WithContext(ctx).
		Model(&model.Link{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

func (d *linkRepoImpl) CheckShortCodeDuplicate(ctx context.Context, tx *gorm.DB, short_code string) (bool, error) {
	if tx == nil {
		tx = d.db
	}
	var count int64
	err := tx.WithContext(ctx).
		Model(&model.Link{}).
		Where("short_code = ?", short_code).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetLinkByID 根据ID查询链接
func (d *linkRepoImpl) GetLinkByID(ctx context.Context, tx *gorm.DB, linkID int64) (*model.Link, error) {
	if tx == nil {
		tx = d.db
	}
	var link model.Link
	err := tx.WithContext(ctx).
		Where("id = ?", linkID).
		First(&link).Error

	if err != nil {
		return nil, err
	}
	return &link, nil
}

// DeleteLinkByID 根据ID删除链接
func (d *linkRepoImpl) DeleteLinkByID(ctx context.Context, tx *gorm.DB, linkID int64) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Where("id = ?", linkID).Delete(&model.Link{}).Error
}
