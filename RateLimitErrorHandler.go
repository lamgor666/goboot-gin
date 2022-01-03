package goboot

import (
	"github.com/lamgor666/goboot-gin/http/error/RateLimitError"
	"github.com/lamgor666/goboot-gin/http/response/HttpError"
)

type rateLimitErrorHandler struct {
}

func (h rateLimitErrorHandler) GetErrorName() string {
	return "builtin.RateLimitError"
}

func (h rateLimitErrorHandler) MatchError(err error) bool {
	if _, ok := err.(RateLimitError.Impl); ok {
		return true
	}

	return false
}

func (h rateLimitErrorHandler) HandleError(_ error) ResponsePayload {
	return HttpError.New(429)
}
