package log

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	commonStr = "%s "
)

type dbLogImpl struct {
	gormLogger.Config
}

func NewDbLogger(config gormLogger.Config) gormLogger.Interface {
	return &dbLogImpl{
		Config: config,
	}
}

func (m *dbLogImpl) LogMode(logLevel gormLogger.LogLevel) gormLogger.Interface {
	l := *m
	l.LogLevel = logLevel
	return &l
}
func (m *dbLogImpl) Info(ctx context.Context, msg string, args ...interface{}) {
	if m.LogLevel >= gormLogger.Info {
		msg = strings.ReplaceAll(msg, "\n", "")
		msg = fmt.Sprintf(commonStr+msg, append([]interface{}{utils.FileWithLineNum()}, args...)...)
		InfoZ(ctx, msg)
	}
}
func (m *dbLogImpl) Warn(ctx context.Context, msg string, args ...interface{}) {
	if m.LogLevel >= gormLogger.Warn {
		msg = strings.ReplaceAll(msg, "\n", "")
		msg = fmt.Sprintf(commonStr+msg, append([]interface{}{utils.FileWithLineNum()}, args...)...)
		WarnZ(ctx, msg)
	}
}
func (m *dbLogImpl) Error(ctx context.Context, msg string, args ...interface{}) {
	if m.LogLevel >= gormLogger.Error {
		msg = strings.ReplaceAll(msg, "\n", "")
		msg = fmt.Sprintf(commonStr+msg, append([]interface{}{utils.FileWithLineNum()}, args...)...)
		ErrorZ(ctx, msg)
	}
}
func (m *dbLogImpl) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if m.LogLevel <= gormLogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && m.LogLevel >= gormLogger.Error:
		sql, rows := fc()
		ErrorZ(ctx, "exec sql raise error, but ignore this error", zap.Any("error", err),
			zap.String("sql", sql), zap.Int64("rows", rows),
			zap.String("sql_used_path", utils.FileWithLineNum()), zap.Duration("elapsed", elapsed))
	case elapsed > m.SlowThreshold && m.SlowThreshold != 0 && m.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", m.SlowThreshold)
		WarnZ(ctx, "exec sql slow", zap.Any("error", err), zap.Duration("elapsed", elapsed),
			zap.String("sql", sql),
			zap.String("slow_log", slowLog), zap.Duration("slow_threshold", m.SlowThreshold), zap.Int64("rows", rows),
			zap.String("sql_used_path", utils.FileWithLineNum()))
	case m.LogLevel == gormLogger.Info:
		sql, rows := fc()
		DebugZ(ctx, "exec sql debug info", zap.String("sql", sql), zap.Int64("rows", rows),
			zap.String("sql_used_path", utils.FileWithLineNum()),
			zap.Duration("elapsed", elapsed))
	}
}
