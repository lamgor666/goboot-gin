package goboot

import (
	"github.com/lamgor666/goboot-common/enum/JwtVerifyErrno"
	"github.com/lamgor666/goboot-gin/http/error/JwtAuthError"
	"github.com/lamgor666/goboot-gin/http/response/JsonResponse"
)

type jwtAuthErrorHandler struct {
}

func (h jwtAuthErrorHandler) GetErrorName() string {
	return "builtin.JwtAuthError"
}

func (h jwtAuthErrorHandler) MatchError(err error) bool {
	if _, ok := err.(JwtAuthError.Impl); ok {
		return true
	}

	return false
}

func (h jwtAuthErrorHandler) HandleError(err error) ResponsePayload {
	ex := err.(JwtAuthError.Impl)
	var code int
	var msg string

	switch ex.Errno() {
	case JwtVerifyErrno.NotFound:
		code = 1001
		msg = "安全令牌缺失"
	case JwtVerifyErrno.Invalid:
		code = 1002
		msg = "不是有效的安全令牌"
	case JwtVerifyErrno.Expired:
		code = 1003
		msg = "安全令牌已失效"
	}

	payload := map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": nil,
	}

	return JsonResponse.New(payload)
}
