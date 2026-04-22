package log

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestLog print request log via http
func RequestLog() gin.HandlerFunc {
	return func(g *gin.Context) {
		traceID := g.Request.Header.Get("Trace-Id")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		ctx := context.WithValue(g.Request.Context(), TraceID, traceID)
		g.Set(string(TraceID), ctx)
		defer ginRecover(ctx, g)
		startTime := time.Now()
		var err error
		var bodyBytes []byte
		path := g.Request.URL.Path
		queryParams := g.Request.URL.RawQuery
		contentType := g.Request.Header.Get("Content-Type")
		if contentType == "" || contentType == "application/json" {
			if g.Request.Body != nil {
				bodyBytes, err = io.ReadAll(g.Request.Body)
				if err != nil {
					WarnZ(ctx, "ReadAll from path failed", zap.Error(err), zap.String("path", path))
				}
				g.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		g.Next()

		nowTimestamp := startTime.UnixNano() / 1e6
		if len(bodyBytes) > 0 {
			InfoZ(ctx, "user request logged",
				zap.Int64("request_receive_time", nowTimestamp),
				zap.String("method", g.Request.Method),
				zap.String("request_path", path),
				zap.String("request_query_params", queryParams),
				zap.ByteString("request_body", bodyBytes),
				zap.Int64("response_time", time.Now().UnixNano()/1e6),
				zap.Duration("elapsed", time.Since(startTime)),
				zap.Int("response_http_status", g.Writer.Status()),
				zap.Int("response_size", g.Writer.Size()),
			)
			return
		}
		InfoZ(ctx, "user request logged",
			zap.Int64("request_receive_time", nowTimestamp),
			zap.String("method", g.Request.Method),
			zap.String("request_path", path),
			zap.String("request_query_params", queryParams),
			zap.Int64("response_time", time.Now().UnixNano()/1e6),
			zap.Duration("elapsed", time.Since(startTime)),
			zap.Int("response_http_status", g.Writer.Status()),
			zap.Int("response_size", g.Writer.Size()),
		)
	}
}

// GinWithCtxHandlerFunc defines the handler used by with context handler
type GinWithCtxHandlerFunc func(context.Context, *gin.Context)

// GinWithCtxHandler handler support context
// how to usage:
// func TestHandler(ctx context.Context, g *gin.Context) {}
// gin.Default().Group("/v1").POST("/test",  GinWithCtxHandler(TestHandler))
func GinWithCtxHandler(f GinWithCtxHandlerFunc) gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx, ok := g.MustGet(string(TraceID)).(context.Context)
		if !ok {
			f(nil, g)
		}
		f(ctx, g)
	}
}

// ginRecover
func ginRecover(ctx context.Context, g *gin.Context) {
	if err := recover(); err != nil {
		// Check for a broken connection, as it is not really a
		// condition that warrants a panic stack trace.
		requestPath := g.Request.URL.Path
		brokenPipe := false
		ne, ok := err.(*net.OpError)
		var se *os.SyscallError
		if ok && errors.As(ne.Err, &se) &&
			(strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
				strings.Contains(strings.ToLower(se.Error()), "connection reset by peer")) {
			brokenPipe = true
		}
		ErrorZ(ctx, "gin handler unknown error",
			zap.String("method", g.Request.Method),
			zap.String("request_path", requestPath),
			zap.Int64("response_time", time.Now().UnixNano()/1e6),
			zap.Int("response_http_status", http.StatusInternalServerError), zap.Any("error", "error"), zap.Stack("stack"))
		if !brokenPipe {
			g.JSON(http.StatusInternalServerError, nil)
		}
		g.Abort()
	}
}
