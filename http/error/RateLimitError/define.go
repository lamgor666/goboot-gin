package RateLimitError

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type Impl struct {
	total      int
	remaining  int
	retryAfter string
}

func New(data map[string]interface{}) Impl {
	var total int

	if n1, ok := data["total"].(int); ok && n1 > 0 {
		total = n1
	}

	var remaining int

	if n1, ok := data["remaining"].(int); ok && n1 > 0 {
		remaining = n1
	}

	var retryAfter string

	if s1, ok := data["retryAfter"].(string); ok && s1 != "" {
		retryAfter = s1
	}

	return Impl{
		total:      total,
		remaining:  remaining,
		retryAfter: retryAfter,
	}
}

func (ex Impl) Error() string {
	return "rate limit exceed"
}

func (ex Impl) Total() int {
	return ex.total
}

func (ex Impl) Remaining() int {
	return ex.remaining
}

func (ex Impl) RetryAfter() string {
	return ex.retryAfter
}

func (ex Impl) AddSpecifyHeaders(ctx *gin.Context) {
	ctx.Header("X-Ratelimit-Limit", fmt.Sprintf("%d", ex.Total()))
	ctx.Header("X-Ratelimit-Remaining", fmt.Sprintf("%d", ex.Remaining()))

	if ex.RetryAfter() != "" {
		ctx.Header("Retry-After", ex.RetryAfter())
	}
}
