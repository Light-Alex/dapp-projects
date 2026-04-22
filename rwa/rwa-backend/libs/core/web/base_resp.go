package web

import (
	"net/http"

	"github.com/AnchoredLabs/rwa-backend/libs/core/error_msg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Response struct {
	Data      interface{} `json:"data"`
	Msg       string      `json:"msg"`
	Code      int         `json:"code"`
	RequestId string      `json:"requestId"`
}

func ResponseOk(data interface{}, g *gin.Context) {
	g.JSON(http.StatusOK, &Response{
		Data:      data,
		Code:      0,
		RequestId: getTraceId(g),
	})
}

func ResponseError(key string, g *gin.Context) {
	errCode, ok := error_msg.GetErrorCode(key)
	if !ok {
		errCode, _ = error_msg.GetErrorCode(error_msg.ErrInternalServerError)
	}
	g.JSON(http.StatusInternalServerError, &Response{
		Msg:       errCode.Msg,
		Code:      errCode.Code,
		RequestId: getTraceId(g),
	})
}

func ResponseUnAuthorizedError(g *gin.Context) {
	g.JSON(http.StatusUnauthorized, &Response{
		Msg:       http.StatusText(http.StatusUnauthorized),
		Code:      http.StatusUnauthorized,
		RequestId: getTraceId(g),
	})
}

func ResponseNotFoundError(msg string, g *gin.Context) {
	if msg == "" {
		msg = http.StatusText(http.StatusNotFound)
	}
	g.JSON(http.StatusNotFound, &Response{
		Msg:       msg,
		Code:      http.StatusNotFound,
		RequestId: getTraceId(g),
	})
}

func getTraceId(_ *gin.Context) string {
	return uuid.New().String()
}
