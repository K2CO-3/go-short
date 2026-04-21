package postgresql

// ==========================================
// AccessLog 相关操作 (Worker Service)
// ==========================================

import (
	"context"
	"go-short/internal/model"
	"go-short/internal/repository"

	"gorm.io/gorm"
)

type accessLogRepoImpl struct {
	db *gorm.DB
}

// NewAccessLogRepository 创建 AccessLogRepository 实例
func NewAccessLogRepository(db *gorm.DB) *accessLogRepoImpl {
	return &accessLogRepoImpl{db: db}
}

// SaveAccessLog 保存访问日志 (由 Worker 调用)
func (d *accessLogRepoImpl) SaveAccessLog(ctx context.Context, tx *gorm.DB, logEntry *model.AccessLog) error {
	if tx == nil {
		tx = d.db
	}
	return tx.WithContext(ctx).Create(logEntry).Error
}

// GetAccessLogs 获取访问日志（支持筛选，按 VisitedAt 倒序）
func (d *accessLogRepoImpl) GetAccessLogs(ctx context.Context, tx *gorm.DB, query repository.AccessLogQuery) ([]model.AccessLog, error) {
	if tx == nil {
		tx = d.db
	}

	dbQuery := tx.WithContext(ctx).
		Model(&model.AccessLog{}).
		Joins("LEFT JOIN links ON links.id = access_logs.link_id")

	if query.StartTime != nil {
		dbQuery = dbQuery.Where("access_logs.visited_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		dbQuery = dbQuery.Where("access_logs.visited_at <= ?", *query.EndTime)
	}
	if query.IPAddress != "" {
		dbQuery = dbQuery.Where("access_logs.ip_address = ?", query.IPAddress)
	}
	if query.ShortCode != "" {
		dbQuery = dbQuery.Where("access_logs.short_code ILIKE ?", "%"+query.ShortCode+"%")
	}
	if query.OriginalURL != "" {
		dbQuery = dbQuery.Where("links.original_url ILIKE ?", "%"+query.OriginalURL+"%")
	}

	var logs []model.AccessLog
	dbQuery = dbQuery.Order("visited_at DESC")
	if query.Limit > 0 {
		dbQuery = dbQuery.Limit(query.Limit)
	}

	err := dbQuery.Find(&logs).Error
	return logs, err
}
