package log

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestConsoleLogger(_ *testing.T) {
	InitLogger(&Conf{EncoderType: ConsoleEncoder})
	InfoZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	InfoZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	ErrorZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	ErrorZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
}

func TestJsonLogger(_ *testing.T) {
	InitLogger(&Conf{EncoderType: JSONEncoder})
	InfoZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	InfoZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	ErrorZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	ErrorZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
}

func TestNoInit(_ *testing.T) {
	InfoZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	InfoZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	ErrorZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
	ErrorZ(context.TODO(), uuid.NewString(), zap.Any("key", "value"))
}

func TestTraceId(_ *testing.T) {
	ctx := context.WithValue(context.TODO(), TraceID, uuid.NewString())
	InitLogger(&Conf{EncoderType: ConsoleEncoder, OutputType: OutputAll, DisableStacktrace: true})
	InfoZ(ctx, uuid.NewString(), zap.Any("key", "value"))
	InfoZ(ctx, uuid.NewString(), zap.Any("key", "value"))
	ctx = context.WithValue(ctx, TraceID, uuid.NewString())
	ErrorZ(ctx, uuid.NewString(), zap.Any("key", "value"))
}

func TestJsonFile(_ *testing.T) {
	ctx := context.WithValue(context.TODO(), TraceID, uuid.NewString())
	InitLogger(&Conf{EncoderType: JSONEncoder, OutputType: OutputAll, DisableStacktrace: true})
	InfoZ(ctx, uuid.NewString(), zap.Any("key", "value"))
	InfoZ(ctx, uuid.NewString(), zap.Any("key", "value"))
	ctx = context.WithValue(ctx, TraceID, uuid.NewString())
	ErrorZ(ctx, uuid.NewString(), zap.Any("key", "value"))
}
