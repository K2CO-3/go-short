package pkg

//目前尚未启用
// logger.go
import (
	"io"
	"os"
	"sync"
)

// 1. 定义日志级别 - 最核心的枚举
type Level int

const (
	LevelDebug Level = iota // 0: 调试
	LevelInfo               // 1: 信息
	LevelWarn               // 2: 警告
	LevelError              // 3: 错误
	LevelFatal              // 4: 致命
)

// 2. 核心日志器结构体 - 这是我们的"主角"
type Logger struct {
	mu       sync.Mutex // 互斥锁，保证并发安全
	out      io.Writer  // 输出目标（控制台、文件等）
	level    Level      // 当前日志级别
	prefix   string     // 日志前缀
	colorful bool       // 是否启用颜色
}

// 3. 创建一个最简单的日志器
func NewLogger() *Logger {
	return &Logger{
		out:      os.Stdout, // 默认输出到控制台
		level:    LevelInfo, // 默认只显示INFO及以上
		colorful: true,      // 默认带颜色
	}
}
