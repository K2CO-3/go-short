package postgresql

import (
	"go-short/internal/model"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresClient 初始化 PostgreSQL 连接
func NewPostgresClient() (*gorm.DB, error) {
	// 1. 从环境变量读取连接字符串 (在 docker-compose.yml 中定义的 DB_DSN)
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// 默认 fallback (仅用于本地非Docker环境调试)
		dsn = "host=localhost user=cmh password=123456 dbname=goshort port=5432 sslmode=disable"
	}

	// 2. 配置 GORM 日志级别 (生产环境建议 Error，开发环境 Info)
	logLevel := logger.Info
	if os.Getenv("APP_ENV") == "production" {
		logLevel = logger.Error
	}

	// 3. 建立连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	// 4. 获取通用数据库对象 sql.DB 以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量
	// 注意：这个值要根据你的 Postgres 容器配置(max_connections)来定，不要超过数据库的上限
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ PostgreSQL connected successfully")

	// 5. 运行自动迁移
	log.Println("🔧 Running database auto migrations...")

	// 要迁移的模型列表
	err = db.AutoMigrate(
		&model.User{},
		&model.Link{},
		&model.AccessLog{},
	)

	if err != nil {
		log.Printf("⚠️  Auto migration warning: %v", err)
		// 不返回错误，让服务继续运行
	} else {
		log.Println("✅ Database migrations completed")
	}
	return db, nil
}
