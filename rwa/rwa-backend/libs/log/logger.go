package log

import (
	"context"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TraceKey string

const TraceID TraceKey = "x-trice-id"
const PriorityTrace = "x-priority"

type AlertPriorityType string
type OutputType string

var onceSink = &sync.Once{}

const (
	// StrLvlDebug debug level
	StrLvlDebug = "debug"
	// StrLvlInfo info level
	StrLvlInfo = "info"
	// StrLvlWarn warning level
	StrLvlWarn = "warn"
	// StrLvlError error level
	StrLvlError = "error"
	// StrLvlFatal fatal level very severe
	StrLvlFatal = "fatal"

	timeFormat     = "2006-01-02 15:04:05.000 MST"
	JSONEncoder    = "json"
	ConsoleEncoder = "console"

	// AlertFirstPriority alert priority
	AlertFirstPriority  AlertPriorityType = "Alert0"
	AlertSecondPriority AlertPriorityType = "Alert1"

	OnlyOutputLogFile OutputType = "file"
	OnlyOutputStdout  OutputType = "stdout"
	OutputAll         OutputType = "all"
)

var (
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
)

type Conf struct {
	DisableStacktrace   bool       `json:"disableStacktrace" yaml:"disableStacktrace" mapstructure:"disableStacktrace"` // default false,production should be true
	FilePath            string     `json:"filePath" yaml:"filePath" mapstructure:"filePath"`
	FileName            string     `json:"fileName" yaml:"fileName" mapstructure:"fileName"`
	MaxSize             int        `json:"maxSize" yaml:"maxSize" mapstructure:"maxSize"`                                     // default 100MB
	MaxAge              int        `json:"maxAge" yaml:"maxAge" mapstructure:"maxAge"`                                        // default 30day
	MaxBackups          int        `json:"maxBackups" yaml:"maxBackups" mapstructure:"maxBackups"`                            // default 100
	RotateIntervalHours int        `json:"rotateIntervalHours" yaml:"rotateIntervalHours" mapstructure:"rotateIntervalHours"` // default 24 hours
	EncoderType         string     `json:"encoderType" yaml:"encoderType" mapstructure:"encoderType"`                         // json,console. default is json
	OutputType          OutputType `json:"outputType" yaml:"outputType" mapstructure:"outputType"`                            // 1: OnlyOutputLogFile 2: OnlyOutputStdout 3: OutputAll  default is 2
	Level               string     `json:"level" yaml:"level" mapstructure:"level"`                                           // debug,info,warn,error,fatal default is info
	EnableColor         bool       `json:"enableColor" yaml:"enableColor" mapstructure:"enableColor"`
	Development         bool       `json:"development" yaml:"development" mapstructure:"development"`
}

func init() {
	InitLogger(&Conf{DisableStacktrace: true})
}

// InitLogger initializes logger library
func InitLogger(conf *Conf) {
	InitLoggerWithSample(nil, conf)
}

func initConf(conf *Conf) *Conf {
	if conf.MaxSize == 0 {
		conf.MaxSize = 100
	}
	if conf.MaxBackups == 0 {
		conf.MaxBackups = 100
	}
	if conf.MaxAge == 0 {
		conf.MaxAge = 30
	}
	if conf.Level == "" {
		conf.Level = StrLvlInfo
	}
	if conf.EncoderType == "" {
		conf.EncoderType = ConsoleEncoder
	}
	if conf.OutputType == "" {
		conf.OutputType = OnlyOutputStdout
	}
	if conf.RotateIntervalHours == 0 {
		conf.RotateIntervalHours = 24
	}
	return conf
}

// InitLoggerWithSample initializes logger library with sampling config
// componentName service name
// disableStacktrace whether to disable stack trace
// encoderName output format encoder:
// 1: json (official)
// 2: console (official)
// 3: errorsFriendlyJson (json adapter for errors package)
// 4: errorsFriendlyConsole (console adapter for errors package)
func InitLoggerWithSample(samplingConfig *zap.SamplingConfig, conf *Conf) {
	initConf(conf)
	var err error
	// reset logger
	Exit()
	zapLogLevel := toZapLevel(conf.Level)
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(timeFormat) + "]")
	}
	customLevelEncoder := zapcore.CapitalLevelEncoder
	customCallerEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}
	if conf.Development {
		customCallerEncoder = zapcore.ShortCallerEncoder
	}
	if conf.EncoderType == JSONEncoder {
		customTimeEncoder = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.Format(timeFormat))
		}
		customLevelEncoder = zapcore.CapitalLevelEncoder
		customCallerEncoder = zapcore.ShortCallerEncoder
	}
	if conf.EnableColor {
		customLevelEncoder = zapcore.CapitalColorLevelEncoder
	}
	encodeCfg := zapcore.EncoderConfig{
		TimeKey:          "t",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		EncodeLevel:      customLevelEncoder,
		EncodeTime:       customTimeEncoder,
		EncodeDuration:   zapcore.MillisDurationEncoder,
		EncodeCaller:     customCallerEncoder,
		ConsoleSeparator: " ",
	}
	atomicZapLeveler := zap.NewAtomicLevelAt(zapLogLevel)
	if samplingConfig == nil {
		samplingConfig = &zap.SamplingConfig{
			Initial:    1000,
			Thereafter: 100,
		}
	}
	cfg := zap.Config{
		Level:             atomicZapLeveler,
		Development:       conf.Development,
		DisableStacktrace: conf.DisableStacktrace,
		Encoding:          conf.EncoderType,
		EncoderConfig:     encodeCfg,
		Sampling:          samplingConfig,
	}
	cfg.OutputPaths, cfg.ErrorOutputPaths = initLogOutput(conf)
	if logger, err = cfg.Build(zap.AddCallerSkip(1)); err != nil {
		panic(err)
	}
	sugarLogger = logger.Sugar()
}

func initLogOutput(conf *Conf) (outputPaths []string, errorOutputPaths []string) {
	if conf.OutputType == OutputAll || conf.OutputType == OnlyOutputLogFile {
		// initFileLogger
		initFileLogger(conf)
		outputPaths = append(outputPaths, "lumberjack:test.log")
		errorOutputPaths = append(errorOutputPaths, "lumberjack:test.log")
		onceSink.Do(func() {
			if err := zap.RegisterSink("lumberjack", func(_ *url.URL) (zap.Sink, error) {
				return lumberjackSink{
					Logger: lumlog,
				}, nil
			}); err != nil {
				panic(err) // support multiple log initializations
			}
		})
	}
	if conf.OutputType == OutputAll || conf.OutputType == OnlyOutputStdout {
		outputPaths = append(outputPaths, "stdout")
		errorOutputPaths = append(errorOutputPaths, "stdout")
	}
	return
}

// ErrorWithPriority error log with alert priority
// first priority will call about developer's phone and second priority about
// second priority will send to slack and third priority about
func ErrorWithPriority(ctx context.Context, alertPriority AlertPriorityType, msg string, fields ...zap.Field) {
	fields = append(fields, []zap.Field{genTraceIDZap(ctx), genPriority(alertPriority)}...)
	logger.Error(msg, fields...)
}

func WarnWithPriority(ctx context.Context, alertPriority AlertPriorityType, msg string, fields ...zap.Field) {
	fields = append(fields, []zap.Field{genTraceIDZap(ctx), genPriority(alertPriority)}...)
	logger.Warn(msg, fields...)
}

func InfoWithPriority(ctx context.Context, alertPriority AlertPriorityType, msg string, fields ...zap.Field) {
	fields = append(fields, []zap.Field{genTraceIDZap(ctx), genPriority(alertPriority)}...)
	logger.Info(msg, fields...)
}

// FatalZ logs a message at FatalLevel. The message includes any fields passed
func FatalZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, []zap.Field{genTraceIDZap(ctx), genPriority(AlertSecondPriority)}...)
	logger.Fatal(msg, fields...)
}

// ErrorZ error log with zap new api, this high performance api, strong advise to use error priority default is  thirdPriority
func ErrorZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, []zap.Field{genTraceIDZap(ctx)}...)
	logger.Error(msg, fields...)
}

// WarnZ warn log with zap new api, this high performance api, strong advise to use
func WarnZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, genTraceIDZap(ctx))
	logger.Warn(msg, fields...)
}

// InfoZ info log with zap new api, this high performance api, strong advise to use
func InfoZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, genTraceIDZap(ctx))
	logger.Info(msg, fields...)
}

// DebugZ debug log with zap new api, this high performance api, strong advise to use
func DebugZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, genTraceIDZap(ctx))
	logger.Debug(msg, fields...)
}

// Logger allow other usage
func Logger() *zap.Logger {
	return logger
}

func Exit() {
	if logger != nil {
		_ = logger.Sync()
	}
	if sugarLogger != nil {
		_ = sugarLogger.Sync()
	}
}

func toZapLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case StrLvlDebug:
		return zapcore.DebugLevel
	case StrLvlInfo:
		return zapcore.InfoLevel
	case StrLvlWarn:
		return zapcore.WarnLevel
	case StrLvlError:
		return zapcore.ErrorLevel
	case StrLvlFatal:
		return zapcore.FatalLevel
	default:
		logger.Warn("level str to zap unknown level", zap.String("level_string", levelStr))
		return zapcore.InfoLevel
	}
}

func genTraceIDZap(ctx context.Context) zap.Field {
	return zap.Any(string(TraceID), ctx.Value(TraceID))
}

func genPriority(errorPriority AlertPriorityType) zap.Field {
	return zap.Any(PriorityTrace, errorPriority)
}
