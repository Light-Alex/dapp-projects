package errors

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// wrap is a helper to construct an *wrapper.
func wrap(err error, format, suffix string, args ...interface{}) Err {
	newErr := Err{
		message:  fmt.Sprintf(format+suffix, args...),
		previous: err,
	}
	newErr.SetLocation(2)
	return newErr
}

// CodeError represents an error on timeout.
type CodeError struct {
	Err
	Code     codes.Code
	HTTPCode uint32
	BizCode  uint32 // custom biz code
}

// GetGRPCCode return grpc code
func (c *CodeError) GetGRPCCode() codes.Code {
	return c.Code
}

// GetHTTPCode get http code
func (c *CodeError) GetHTTPCode() uint32 {
	return c.HTTPCode
}

// GetBizCode get biz code
func (c *CodeError) GetBizCode() uint32 {
	return c.BizCode
}

// NewCodeError new code error
func NewCodeError(code codes.Code, httpCode, bizCode uint32) error {
	return &CodeError{Err: wrap(nil, "", ""), Code: code, HTTPCode: httpCode, BizCode: bizCode}
}

// NewCodeErrorf new code errorf
func NewCodeErrorf(code codes.Code, httpCode, bizCode uint32, format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, "", args...), Code: code, HTTPCode: httpCode, BizCode: bizCode}
}

// NewHTTPErrorf new http errorf
func NewHTTPErrorf(httpCode, bizCode uint32, format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, "", args...), HTTPCode: httpCode, BizCode: bizCode}
}

// NewGRPCError new code error
func NewGRPCError(code codes.Code, err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: code}
}

// NewHTTPError new code error
func NewHTTPError(httpCode, bizCode uint32, err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), HTTPCode: httpCode, BizCode: bizCode}
}

// NotValidf returns an error which satisfies IsNotValid().
func NotValidf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusBadRequest), args...), Code: codes.InvalidArgument, HTTPCode: http.StatusBadRequest, BizCode: 12000003}
}

// NewNotValid returns an error which wraps err and satisfies IsNotValid().
func NewNotValid(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.InvalidArgument, HTTPCode: http.StatusBadRequest, BizCode: 12000003}
}

// IsNotValid is not valid error
func IsNotValid(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.InvalidArgument
	}
	return false
}

// Timeoutf returns an error which satisfies IsTimeout().
func Timeoutf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusRequestTimeout), args...), Code: codes.DeadlineExceeded, HTTPCode: http.StatusRequestTimeout, BizCode: 12000004}
}

// NewTimeout returns an error which wraps err that satisfies
func NewTimeout(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.DeadlineExceeded, HTTPCode: http.StatusRequestTimeout, BizCode: 12000004}
}

// IsTimeout check timeout error
func IsTimeout(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.DeadlineExceeded
	}
	return false
}

// NotFoundf returns an error which satisfies IsNotFound().
func NotFoundf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusNotFound), args...), Code: codes.NotFound, HTTPCode: http.StatusNotFound, BizCode: 12000005}
}

// NewNotFound returns an error which wraps err that satisfies
func NewNotFound(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.NotFound, HTTPCode: http.StatusNotFound, BizCode: 12000005}
}

// IsNotFound is not Fund
func IsNotFound(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.NotFound
	}
	return false
}

// AlreadyExistsf returns an error which satisfies
func AlreadyExistsf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusConflict), args...), Code: codes.AlreadyExists, HTTPCode: http.StatusConflict, BizCode: 12000006}
}

// NewAlreadyExists returns an error which wraps err and satisfies
func NewAlreadyExists(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.AlreadyExists, HTTPCode: http.StatusConflict, BizCode: 12000006}
}

// IsAlreadyExists is already exists
func IsAlreadyExists(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.AlreadyExists
	}
	return false
}

// Forbiddenf returns an error which satistifes IsForbidden()
func Forbiddenf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusForbidden), args...), Code: codes.PermissionDenied, HTTPCode: http.StatusForbidden, BizCode: 12000007}
}

// NewForbidden returns an error which wraps err that satisfies
func NewForbidden(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.PermissionDenied, HTTPCode: http.StatusForbidden, BizCode: 12000007}
}

// IsForbidden is forbidden error
func IsForbidden(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.PermissionDenied
	}
	return false
}

// ResourceExhaustedf returns an error which satisfaction IsResourceExhausted()
func ResourceExhaustedf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusGone), args...), Code: codes.ResourceExhausted, HTTPCode: http.StatusGone, BizCode: 12000008}
}

// NewResourceExhausted returns an error which wraps err that satisfies
func NewResourceExhausted(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.ResourceExhausted, HTTPCode: http.StatusGone, BizCode: 12000008}
}

// IsResourceExhausted is resource exhausted error
func IsResourceExhausted(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.ResourceExhausted
	}
	return false
}

// FailedPreconditionf returns an error which satisfaction IsFailedPrecondition()
func FailedPreconditionf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusPreconditionFailed), args...), Code: codes.FailedPrecondition, HTTPCode: http.StatusPreconditionFailed, BizCode: 12000009}
}

// NewFailedPrecondition returns an error which wraps err that satisfies
func NewFailedPrecondition(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.FailedPrecondition, HTTPCode: http.StatusPreconditionFailed, BizCode: 12000009}
}

// IsFailedPrecondition is failed precondition errors
func IsFailedPrecondition(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.FailedPrecondition
	}
	return false
}

// Abortedf returns an error which satisfaction IsAborted()
func Abortedf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusInternalServerError), args...), Code: codes.Aborted, HTTPCode: http.StatusInternalServerError, BizCode: 12000010}
}

// NewAborted returns an error which wraps err that satisfies
func NewAborted(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.Aborted, HTTPCode: http.StatusInternalServerError, BizCode: 12000010}
}

// IsAborted is aborted error
func IsAborted(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.Aborted
	}
	return false
}

// NotImplementedf returns an error which satisfies IsNotImplemented().
func NotImplementedf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusNotImplemented), args...), Code: codes.Unimplemented, HTTPCode: http.StatusNotImplemented, BizCode: 12000012}
}

// NewNotImplemented returns an error which wraps err and satisfies
func NewNotImplemented(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.Unimplemented, HTTPCode: http.StatusNotImplemented, BizCode: 12000012}
}

// IsNotImplemented is not implemented
func IsNotImplemented(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.Unimplemented
	}
	return false
}

// Intervalf returns an error which internal server error
func Intervalf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusInternalServerError), args...), Code: codes.Internal, HTTPCode: http.StatusInternalServerError, BizCode: 12000013}
}

// IsInterval is internal error
func IsInterval(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.Internal
	}
	return false
}

// Unavailablef returns an error which server unavailable
func Unavailablef(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusServiceUnavailable), args...), Code: codes.Unavailable, HTTPCode: http.StatusServiceUnavailable, BizCode: 12000014}
}

// NewUnavailable returns an error which server unavailable
func NewUnavailable(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.Unavailable, HTTPCode: http.StatusServiceUnavailable, BizCode: 12000014}
}

// IsUnavailable is unavailable error
func IsUnavailable(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.Unavailable
	}
	return false
}

// Unauthorizedf returns an error which satisfies IsUnauthorized().
func Unauthorizedf(format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, " "+http.StatusText(http.StatusUnauthorized), args...), Code: codes.Unauthenticated, HTTPCode: http.StatusUnauthorized, BizCode: 12000016}
}

// NewUnauthorized returns an error which wraps err and satisfies
func NewUnauthorized(err error, msg string) error {
	return &CodeError{Err: wrap(err, msg, ""), Code: codes.Unauthenticated, HTTPCode: http.StatusUnauthorized, BizCode: 12000016}
}

// IsUnauthorized is unauthorized
func IsUnauthorized(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.Code == codes.Unauthenticated
	}
	return false
}

// ToGRPCStatus to grpc status error
func ToGRPCStatus(err error) *status.Status {
	err = Cause(err)
	if err == nil {
		return nil
	}
	var inner *CodeError
	ok := errors.As(err, &inner)
	if ok {
		if inner.Code == codes.OK && inner.BizCode != 0 {
			inner.Code = codes.Unknown
		}
		st := status.New(inner.Code, err.Error())
		if inner.BizCode != 0 {
			st, _ = st.WithDetails(&BizErrorCode{Code: inner.BizCode})
		}
		return st
	}
	st, _ := status.FromError(err)
	return st
}

// ToGRPCReturnError generate grpc api return error
func ToGRPCReturnError(err error) error {
	st := ToGRPCStatus(err)
	if st == nil {
		return nil
	}
	return st.Err()
}

// GRPCErrToError grpc error to code error
func GRPCErrToError(err error) error {
	if err == nil {
		return nil
	}
	st, _ := status.FromError(err)
	httpCode := grpcCodeToHttpCode[st.Code()]
	codeErr := &CodeError{Err: wrap(nil, err.Error(), ""), Code: st.Code(), HTTPCode: httpCode}
	details := st.Details()
	for _, detail := range details {
		if bizErrorCode, ok := detail.(*BizErrorCode); ok {
			codeErr.BizCode = bizErrorCode.Code
		}
	}
	return codeErr
}

// NewBizCodeError new biz code error with biz code and translated message
// Suggest using NewCodeError for more detailed code
func NewBizCodeError(bizCode uint32) error {
	return &CodeError{Err: wrap(nil, "bizCode: %d", "", bizCode), Code: codes.OK, HTTPCode: http.StatusOK, BizCode: bizCode}
}

// NewBizCodeErrorf new biz code error with biz code and custom message
// Suggest using NewCodeErrorf for more detailed code
func NewBizCodeErrorf(bizCode uint32, format string, args ...interface{}) error {
	return &CodeError{Err: wrap(nil, format, "", args...), Code: codes.OK, HTTPCode: http.StatusOK, BizCode: bizCode}
}

// // FromNamerErr from namer error to code error
// func FromNamerErr(err error) error {
// 	if err == nil {
// 		return nil
// 	}
// 	desc, ok := err.(*namer.ErrDesc)
// 	if !ok {
// 		return NewInterval(err, "")
// 	}
// 	return NewBizCodeErrorf(desc.Code, desc.Message)
// }

// IsBizCodeError is biz code error
func IsBizCodeError(err error, bizCode uint32) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		return innerErr.BizCode == bizCode
	}
	return false
}

// IsAnyBizCodeErr is any biz code error
func IsAnyBizCodeErr(err error) bool {
	err = Cause(err)
	var innerErr *CodeError
	if errors.As(err, &innerErr) {
		if innerErr.Code == codes.OK && innerErr.HTTPCode == http.StatusOK && innerErr.BizCode > 0 {
			return true
		}
	}
	return false
}
