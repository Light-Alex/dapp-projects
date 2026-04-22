package log

import (
	"context"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var lumlog *lumberjack.Logger

type lumberjackSink struct {
	*lumberjack.Logger
}

var onceLum = &sync.Once{}

const logFileAfterFix = ".log"

// Sync implements zap.Sink. The remaining methods are implemented
// by the embedded *lumberjack.Logger.
func (lumberjackSink) Sync() error { return nil }

func initFileLogger(conf *Conf) {
	fileName := "logs/app"
	if conf.FilePath != "" && conf.FileName != "" {
		fileName = conf.FilePath + "/" + conf.FileName
	}
	onceLum.Do(func() {
		lumlog = &lumberjack.Logger{
			Filename:   fileName + logFileAfterFix,
			MaxSize:    conf.MaxSize, // megabytes
			LocalTime:  true,
			MaxBackups: conf.MaxBackups,
			MaxAge:     conf.MaxAge, // days
		}
		go func() {
			intervalSec := int64(conf.RotateIntervalHours) * 3600
			now := time.Now()
			passed := int64(now.Hour()*3600 + now.Minute()*60 + now.Second())
			leftTs := intervalSec - (passed % intervalSec) - 1

			timer := time.NewTimer(time.Duration(leftTs) * time.Second)
			for range timer.C {
				if err := lumlog.Rotate(); err != nil {
					WarnZ(context.Background(), "log rotate error")
				}
				timer.Reset(time.Duration(intervalSec) * time.Second)
			}
		}()
	})
}
