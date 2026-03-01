package postgresql

// ==========================================
// AccessLog 相关操作 (Worker Service)
// ==========================================

import (
	"context"
	"go-short/internal/model"

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

// GetRecentAccessLogs 获取最近 N 条访问日志（按 VisitedAt 倒序）
func (d *accessLogRepoImpl) GetRecentAccessLogs(ctx context.Context, tx *gorm.DB, limit int) ([]model.AccessLog, error) {
	if tx == nil {
		tx = d.db
	}
	var logs []model.AccessLog
	err := tx.WithContext(ctx).Model(&model.AccessLog{}).
		Order("visited_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
