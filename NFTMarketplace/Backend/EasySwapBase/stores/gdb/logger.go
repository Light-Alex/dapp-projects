package gdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/go-stack/stack"
	"gorm.io/gorm/logger"

	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
)

const (
	traceInfo  = "%s [%.3fms] [rows:%v] %s"
	traceWarn  = "%s %s [%.3fms] [rows:%v] %s"
	traceError = "%s %s [%.3fms] [rows:%v] %s"
)

// Logger 日志记录器
type Logger struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

// NewLogger 新建日志记录器
func NewLogger(logLevel logger.LogLevel, slowThreshold time.Duration) *Logger {
	return &Logger{LogLevel: logLevel, SlowThreshold: slowThreshold}
}

// LogMode 设置日志记录模式
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info Info日志记录
func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		xzap.WithContext(ctx).Info("DB", zap.String("content", fmt.Sprintf(msg, data...)))
	}
}

// Warn Warn日志记录
func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		xzap.WithContext(ctx).Warn("DB", zap.String("content", fmt.Sprintf(msg, data...)))
	}
}

// Error Error日志记录
func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		xzap.WithContext(ctx).Error("DB", zap.String("content", fmt.Sprintf(msg, data...)))
	}
}

// Trace Trace日志记录
func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// 如果日志级别大于静默模式
	if l.LogLevel > logger.Silent {
		// 计算经过的时间
		elapsed := time.Since(begin)
		// 根据条件执行不同的日志记录操作
		switch {
		// 如果发生错误且日志级别大于等于错误级别
		case err != nil && l.LogLevel >= logger.Error:
			// 执行fc函数获取SQL和行数
			sql, rows := fc()
			// 如果行数为-1，表示无行数
			if rows == -1 {
				// 记录错误日志
				l.Error(ctx, traceError, FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				// 记录错误日志，包含行数
				l.Error(ctx, traceError, FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		// 如果经过的时间大于慢查询阈值且阈值不为0且日志级别大于等于警告级别
		case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
			// 执行fc函数获取SQL和行数
			sql, rows := fc()
			// 构造慢查询日志信息
			slowLog := fmt.Sprintf("Slow SQL Greater Than %v", l.SlowThreshold)
			// 如果行数为-1，表示无行数
			if rows == -1 {
				// 记录慢查询警告日志
				// log.Slowf(traceWarn, FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
				l.Warn(ctx, traceWarn, FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				// 记录慢查询警告日志，包含行数
				// log.Slowf(traceWarn, FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
				l.Warn(ctx, traceWarn, FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		// 如果日志级别为信息级别
		case l.LogLevel == logger.Info:
			// 执行fc函数获取SQL和行数
			sql, rows := fc()
			// 如果行数为-1，表示无行数
			if rows == -1 {
				// 记录信息日志
				l.Info(ctx, traceInfo, FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				// 记录信息日志，包含行数
				l.Info(ctx, traceInfo, FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		}
	}
}

// FileWithLineNum 获取调用堆栈信息
func FileWithLineNum() string {
	// 获取堆栈跟踪信息，并裁剪掉当前函数及其调用者的堆栈信息
	cs := stack.Trace().TrimBelow(stack.Caller(2)).TrimRuntime()

	// 遍历裁剪后的堆栈跟踪信息
	for _, c := range cs {
		// 将堆栈信息格式化为字符串
		s := fmt.Sprintf("%+v", c)
		// 判断堆栈信息中是否不包含 "gorm.io/gorm"
		if !strings.Contains(s, "gorm.io/gorm") {
			// 如果不包含，则返回该堆栈信息
			return s
		}
	}

	// 如果所有堆栈信息都包含 "gorm.io/gorm"，则返回空字符串
	return ""
}
