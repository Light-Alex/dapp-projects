package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReturnData struct {
	Code    uint32 `json:"code"`
	Data    interface{}
	Message string `json:"message"`
}

const (
	OkBizCode uint32 = 0
)

// ResponseErr response error, if err is not code error type, default return http.StatusInternalServerError
func ResponseErr(g *gin.Context, err error) {
	if !responseByErr(g, err) {
		Response(g, http.StatusInternalServerError, OkBizCode, nil, err.Error())
	}
}

// ResponseBad response error, if err is not code error type, default return http.StatusBadRequest
func ResponseBad(g *gin.Context, err error) {
	if !responseByErr(g, err) {
		Response(g, http.StatusBadRequest, OkBizCode, nil, err.Error())
	}
}

// ResponseOk response ok
func ResponseOk(g *gin.Context, data interface{}) {
	Response(g, http.StatusOK, OkBizCode, data, "")
}

// ResponseOkWithMsg response ok with custom message
func ResponseOkWithMsg(g *gin.Context, data interface{}, msg string) {
	Response(g, http.StatusOK, OkBizCode, data, msg)
}

// Response response json, if the above api doesn't satisfy your demands, should be used
func Response(g *gin.Context, httpCode, errCode uint32, data interface{}, message string) {
	g.JSON(int(httpCode), gin.H{
		"code":    errCode,
		"message": message,
		"data":    data,
	})
}

func responseByErr(g *gin.Context, err error) bool {
	if err == nil {
		Response(g, http.StatusOK, OkBizCode, nil, "")
		return true
	}
	message := err.Error()
	err = Cause(err)
	var inner *CodeError
	if errors.As(err, &inner) {
		var errCode uint32
		// if custom biz code, should be used
		if inner.BizCode != OkBizCode {
			errCode = inner.BizCode
			message = err.Error()
		}
		Response(g, inner.HTTPCode, errCode, nil, message)
		return true
	}
	return false
}
