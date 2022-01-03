package goboot

import (
	"github.com/lamgor666/goboot-common/util/jsonx"
	"github.com/lamgor666/goboot-gin/http/error/ValidateError"
	"github.com/lamgor666/goboot-gin/http/response/JsonResponse"
)

type validateErrorHandler struct {
}

func (h validateErrorHandler) GetErrorName() string {
	return "builtin.ValidateError"
}

func (h validateErrorHandler) MatchError(err error) bool {
	if _, ok := err.(ValidateError.Impl); ok {
		return true
	}

	return false
}

func (h validateErrorHandler) HandleError(err error) ResponsePayload {
	ex := err.(ValidateError.Impl)
	code := 1006
	var msg string

	if ex.Failfast() {
		msg = ex.Error()
	} else {
		msg = jsonx.ToJson(ex.ValidateErrors())
	}

	payload := map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": nil,
	}

	return JsonResponse.New(payload)
}
