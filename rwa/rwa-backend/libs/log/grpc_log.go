package log

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientLogTraceInterceptorWrapper grpc unary client log trace interceptor
// example:
// grpc.Dial(host,grpc.WithInsecure(), grpc.WithUnaryInterceptor(UnaryClientLogTraceInterceptorWrapper("demo_service")))
func UnaryClientLogTraceInterceptorWrapper(fromServiceName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if traceValue := ctx.Value(TraceID); traceValue != nil {
			traceID, ok := traceValue.(string)
			if !ok {
				traceID = uuid.NewString()
			}
			ctx = metadata.AppendToOutgoingContext(ctx, string(TraceID), traceID)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "from_service_name", fromServiceName)
		startTime := time.Now()
		InfoZ(ctx, "client call grpc begin", zap.Reflect("grpc_req", req), zap.String("method", method), zap.Time("call_start_time", startTime))
		result := invoker(ctx, method, req, reply, cc, opts...)
		InfoZ(ctx, "client call grpc api response",
			zap.Duration("elapsed", time.Since(startTime)),
			zap.String("method", method),
			zap.Any("grpc_req", req),
			zap.String("from_service_name", fromServiceName))
		return result
	}
}

// UnaryServerLogTraceInterceptor grpc unary server log trace interceptor
// example:
// s := grpc.NewServer(grpc.UnaryInterceptor(UnaryServerLogTraceInterceptor))
// xxx.RegisterRouteGuideServer(s, &server{})
func UnaryServerLogTraceInterceptor(ctx context.Context, req interface{}, serverInfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fromServiceName := "unknown"
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if requestIDs := md.Get(string(TraceID)); len(requestIDs) > 0 {
			ctx = context.WithValue(ctx, TraceID, requestIDs[0])
		} else {
			ctx = context.WithValue(ctx, TraceID, uuid.NewString())
		}

		if values := md.Get("from_service_name"); len(values) > 0 {
			fromServiceName = values[0]
		}
	}
	startTime := time.Now()
	InfoZ(ctx, "grpc server handle begin", zap.Reflect("grpc_req", req), zap.String("method", serverInfo.FullMethod), zap.Time("handle_start_time", startTime))
	result, err := handler(ctx, req)
	InfoZ(ctx, "grpc server handle request",
		zap.Duration("elapsed", time.Since(startTime)),
		zap.Error(err), zap.String("method", serverInfo.FullMethod),
		zap.Reflect("grpc_req", req),
		zap.Reflect("from_service_name", fromServiceName))
	return result, err
}
