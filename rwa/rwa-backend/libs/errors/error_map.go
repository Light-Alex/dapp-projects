package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

var grpcCodeToHttpCode = map[codes.Code]uint32{
	codes.OK:                 http.StatusOK,
	codes.Canceled:           http.StatusRequestTimeout,
	codes.Unknown:            http.StatusInternalServerError,
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.DeadlineExceeded:   http.StatusRequestTimeout,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.ResourceExhausted:  http.StatusGone,
	codes.FailedPrecondition: http.StatusPreconditionFailed,
	codes.Aborted:            http.StatusForbidden,
	codes.OutOfRange:         http.StatusPreconditionFailed,
	codes.Unimplemented:      http.StatusNotImplemented,
	codes.Internal:           http.StatusInternalServerError,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.DataLoss:           http.StatusGone,
	codes.Unauthenticated:    http.StatusUnauthorized,
}
