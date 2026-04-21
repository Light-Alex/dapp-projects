package xgrpc

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"time"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/ProjectsTask/EasySwapBase/errcode"
	logging "github.com/ProjectsTask/EasySwapBase/logger"
	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
)

// PayloadUnaryServerInterceptor 一元服务器拦截器，用于记录服务端请求和响应
func PayloadUnaryServerInterceptor(zapLogger *xzap.ZapLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		// 评估服务端选项
		o := evaluateServerOpt(zapLogger.Opts)
		// 记录开始时间
		startTime := time.Now()
		// 创建新的日志上下文
		newCtx := newServerLoggerCaller(ctx, zapLogger.Logger, info.FullMethod, startTime, o.TimestampFormat())
		// 捕获panic
		defer func() {
			if r := recover(); r != nil {
				// 从panic中恢复
				err = recoverFrom(newCtx, r, GrpcRecoveryHandlerFunc)
			}
		}()

		// 调用原始的服务端处理函数
		resp, err := handler(newCtx, req)
		// 判断是否应该记录日志
		if !o.ShouldLog(info.FullMethod, err) {
			return resp, err
		}

		// 将请求转换为日志字段
		fields := protoMessageToFields(req, "grpc.request")
		// 如果没有错误，将响应也转换为日志字段
		if err == nil {
			fields = append(fields, protoMessageToFields(resp, "grpc.response")...)
		}
		// 获取错误代码
		code := o.CodeFunc(err)
		// 获取日志级别
		level := o.LevelFunc(code)
		// 记录日志
		o.MessageFunc(newCtx, "info", level, err, append(fields, o.DurationFunc(time.Since(startTime))))

		// 包装错误
		return resp, wrapErr(err)
	}
}

// PayloadStreamServerInterceptor 流拦截器，用于记录服务端请求和响应
func PayloadStreamServerInterceptor(zapLogger *xzap.ZapLogger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		// 解析服务端选项
		o := evaluateServerOpt(zapLogger.Opts)
		// 记录开始时间
		startTime := time.Now()
		// 创建新的上下文并附加日志记录器
		ctx := newServerLoggerCaller(stream.Context(), zapLogger.Logger, info.FullMethod, startTime, o.TimestampFormat())
		// 包装ServerStream并附加上下文
		wrapped := &wrappedServerStream{ServerStream: stream, wrappedContext: ctx}

		// 延迟执行的函数，用于捕获panic并处理
		defer func() {
			if r := recover(); r != nil {
				// 从panic中恢复并处理错误
				err = recoverFrom(stream.Context(), r, GrpcRecoveryHandlerFunc)
			}
		}()

		// 调用原始处理器
		err = handler(srv, wrapped)

		// 如果不需要记录日志，则直接返回错误
		if !o.ShouldLog(info.FullMethod, err) {
			return err
		}

		// 获取错误代码
		code := o.CodeFunc(err)
		// 获取日志级别
		level := o.LevelFunc(code)
		// 记录日志信息
		o.MessageFunc(ctx, "info", level, err, []zap.Field{o.DurationFunc(time.Since(startTime))})

		// 包装错误并返回
		return wrapErr(err)
	}
}

// wrappedServerStream 包装后的服务端流对象
type wrappedServerStream struct {
	grpc.ServerStream
	wrappedContext context.Context
}

// SendMsg 发送消息
func (l *wrappedServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	if err == nil {
		addFields(l.Context(), protoMessageToFields(m, "grpc.response")...)
	}

	return wrapErr(err)
}

// RecvMsg 接收消息
func (l *wrappedServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	if err == nil {
		addFields(l.Context(), protoMessageToFields(m, "grpc.request")...)
	}

	return wrapErr(err)
}

// Context 返回封装的上下文
func (l *wrappedServerStream) Context() context.Context {
	return l.wrappedContext
}

func evaluateServerOpt(opts []xzap.Option) *xzap.Options {
	optCopy := &xzap.Options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}

	return optCopy
}

// newServerLoggerCaller 创建一个带有日志记录器调用者信息的上下文
//
// 参数:
// ctx: 当前上下文
// logger: zap日志记录器
// methodString: gRPC方法字符串
// start: 请求开始时间
// timestampFormat: 时间戳格式
//
// 返回值:
// 带有日志记录器调用者信息的上下文
func newServerLoggerCaller(ctx context.Context, logger *zap.Logger, methodString string, start time.Time, timestampFormat string) context.Context {
	// 初始化字段切片
	var fields []zapcore.Field
	// 添加开始时间字段
	fields = append(fields, zap.String("grpc.start_time", start.Format(timestampFormat)))

	// 如果上下文中存在截止时间
	if d, ok := ctx.Deadline(); ok {
		// 添加请求截止时间字段
		fields = append(fields, zap.String("grpc.request.deadline", d.Format(timestampFormat)))
	}

	// 如果上下文中存在对等方信息
	if p, ok := peer.FromContext(ctx); ok {
		// 添加对等方地址字段
		fields = append(fields, zap.String("grpc.address", p.Addr.String()))
	}

	// 返回带有日志记录器的上下文
	return xzap.ToContext(ctx, logger.With(append(fields, serverCallFields(methodString)...)...))
}

// serverCallFields 服务端日志fields
func serverCallFields(methodString string) []zapcore.Field {
	service := path.Dir(methodString)[1:]
	method := path.Base(methodString)
	return []zapcore.Field{
		SystemField,
		ServerField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}

// ------------------------------------- 客户端 ----------------------------------

// PayloadUnaryClientInterceptor 一元拦截器，用于记录客户端端请求和响应
func PayloadUnaryClientInterceptor(zapLogger *xzap.ZapLogger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		// 评估客户端选项
		o := evaluateClientOpt(zapLogger.Opts)
		startTime := time.Now()
		// 创建新的上下文
		newCtx := newClientLoggerCaller(ctx, zapLogger.Logger, method, startTime, o.TimestampFormat())
		defer func() {
			// 恢复异常处理
			if r := recover(); r != nil {
				err = recoverFrom(newCtx, r, GrpcRecoveryHandlerFunc)
			}
		}()

		// 调用invoker执行请求
		err = invoker(newCtx, method, req, resp, cc, opts...)
		if !o.ShouldLog(method, err) {
			return err
		}

		// 将请求对象转换为日志字段
		fields := protoMessageToFields(req, "grpc.request")
		if err == nil {
			// 如果请求成功，将响应对象也转换为日志字段并添加到字段列表中
			fields = append(fields, protoMessageToFields(resp, "grpc.response")...)
		}

		// 获取日志级别
		level := o.LevelFunc(o.CodeFunc(err))
		// 计算请求耗时
		duration := o.DurationFunc(time.Since(startTime))
		// 将耗时添加到字段列表中
		fields = append(fields, duration)
		// 记录日志
		o.MessageFunc(newCtx, "info", level, err, fields)

		return err
	}
}

// PayloadStreamClientInterceptor 是一个 gRPC 流客户端拦截器函数，用于拦截 gRPC 流客户端的调用。
// 参数 zapLogger 是一个指向 xzap.ZapLogger 类型的指针，用于日志记录。
// 返回值是一个 grpc.StreamClientInterceptor 类型的函数，用于拦截 gRPC 流客户端的调用。
func PayloadStreamClientInterceptor(zapLogger *xzap.ZapLogger) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (_ grpc.ClientStream, err error) {
		// 评估客户端选项
		o := evaluateClientOpt(zapLogger.Opts)
		// 记录开始时间
		startTime := time.Now()
		// 创建新的上下文
		newCtx := newClientLoggerCaller(ctx, zapLogger.Logger, method, startTime, o.TimestampFormat())
		// 延迟执行恢复操作
		defer func() {
			if r := recover(); r != nil {
				// 从恢复操作中获取错误
				err = recoverFrom(newCtx, r, GrpcRecoveryHandlerFunc)
			}
		}()

		// 创建客户端流
		clientStream, err := streamer(newCtx, desc, cc, method, opts...)
		// 如果不需要记录日志
		if !o.ShouldLog(method, err) {
			if err != nil {
				// 如果发生错误，返回错误
				return nil, err
			}

			// 返回包装后的客户端流
			return &wrappedClientStream{
				ClientStream:   clientStream,
				wrappedContext: newCtx,
			}, err
		}

		// 获取日志级别
		level := o.LevelFunc(o.CodeFunc(err))
		var fields []zap.Field
		// 计算持续时间
		duration := o.DurationFunc(time.Since(startTime))
		fields = append(fields, duration)
		// 记录日志信息
		o.MessageFunc(newCtx, "info", level, err, fields)

		// 返回包装后的客户端流
		return &wrappedClientStream{
			ClientStream:   clientStream,
			wrappedContext: newCtx,
		}, nil
	}
}

// wrappedClientStream 包装后的客户端流对象
type wrappedClientStream struct {
	grpc.ClientStream
	wrappedContext context.Context
}

// SendMsg 发送消息
func (l *wrappedClientStream) SendMsg(m interface{}) error {
	// 调用客户端流的 SendMsg 方法发送消息
	err := l.ClientStream.SendMsg(m)

	// 如果发送消息没有错误
	if err == nil {
		// 将消息添加到上下文中
		addFields(l.Context(), protoMessageToFields(m, "grpc.request")...)
	}

	// 返回包装后的错误
	return wrapErr(err)
}

// RecvMsg 接收消息
func (l *wrappedClientStream) RecvMsg(m interface{}) error {
	// 从客户端流接收消息
	err := l.ClientStream.RecvMsg(m)
	if err == nil {
		// 如果没有错误，将消息字段添加到上下文中
		// 添加字段到上下文中
		addFields(l.Context(), protoMessageToFields(m, "grpc.response")...)
	}

	// 包装并返回错误
	return wrapErr(err)
}

// Context 返回封装的上下文, 用于覆盖 grpc.ServerStream.Context()
func (l *wrappedClientStream) Context() context.Context {
	return l.wrappedContext
}

type protoMessageObject struct {
	pb proto.Message
}

// MarshalLogObject 序列化成日志对象
func (j *protoMessageObject) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	return oe.AddReflected("content", j)
}

// MarshalJSON 序列化成json
func (j *protoMessageObject) MarshalJSON() ([]byte, error) {
	// 创建一个bytes.Buffer对象
	b := &bytes.Buffer{}

	// 使用JsonPbMarshaller将protoMessageObject对象序列化为JSON格式并写入Buffer中
	if err := JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		// 如果序列化失败，则返回错误信息
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}

	// 返回Buffer中的字节数据
	return b.Bytes(), nil
}

// protoMessageToFields 将message序列化成json，并写入存储
func protoMessageToFields(pbMsg interface{}, key string) []zap.Field {
	var fields []zap.Field // 定义一个zap.Field类型的切片
	// 判断pbMsg是否可以转换为proto.Message类型
	if p, ok := pbMsg.(proto.Message); ok {
		// 将proto.Message类型的对象转换为zap.Object类型，并添加到fields切片中
		fields = append(fields, zap.Object(key, &protoMessageObject{pb: p}))
	}

	return fields
}

// recoverFrom 恐慌处理
func recoverFrom(ctx context.Context, p interface{}, r logging.RecoveryHandlerContextFunc) error {
	if r == nil {
		return status.Errorf(codes.Internal, "%v", p)
	}
	return r(ctx, p)
}

// wrapErr 返回gRPC状态码包装后的业务错误
func wrapErr(err error) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case interface{ GRPCStatus() *status.Status }:
		return e.GRPCStatus().Err()
	case *errcode.Err:
		return status.Error(codes.Code(e.Code()), e.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}

func evaluateClientOpt(opts []xzap.Option) *xzap.Options {
	optCopy := &xzap.Options{}
	*optCopy = *defaultClientOptions
	for _, o := range opts {
		o(optCopy)
	}

	return optCopy
}

// newClientLoggerCaller 新建客户端
func newClientLoggerCaller(ctx context.Context, logger *zap.Logger, methodString string, start time.Time, timestampFormat string) context.Context {
	// 初始化字段切片
	var fields []zapcore.Field
	// 添加grpc.start_time字段
	fields = append(fields, zap.String("grpc.start_time", start.Format(timestampFormat)))
	// 判断ctx是否有截止时间
	if d, ok := ctx.Deadline(); ok {
		// 添加grpc.request.deadline字段
		fields = append(fields, zap.String("grpc.request.deadline", d.Format(timestampFormat)))
	}

	// 判断ctx是否有对等方信息
	if p, ok := peer.FromContext(ctx); ok {
		// 添加grpc.address字段
		fields = append(fields, zap.String("grpc.address", p.Addr.String()))
	}

	// 将logger与字段组合并返回新的context
	return xzap.ToContext(ctx, logger.With(append(fields, clientLoggerFields(methodString)...)...))
}

// clientLoggerFields 客户端日志fields
func clientLoggerFields(methodString string) []zapcore.Field {
	// 获取方法字符串的路径目录部分，并去除目录部分的第一个字符（通常是'/'）
	service := path.Dir(methodString)[1:]
	// 获取方法字符串的文件名部分
	method := path.Base(methodString)
	return []zapcore.Field{
		// 系统字段
		SystemField,
		// 客户端字段
		ClientField,
		// gRPC服务字段
		zap.String("grpc.service", service),
		// gRPC方法字段
		zap.String("grpc.method", method),
	}
}

// addFields 添加zap Field 到日志中
func addFields(ctx context.Context, fields ...zap.Field) {
	// 获取与上下文关联的日志记录器
	l := xzap.WithContext(ctx)
	// 使用给定的字段更新日志记录器
	l.WithField(fields...)
}
