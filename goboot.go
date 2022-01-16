package goboot

import (
	"github.com/gin-gonic/gin"
	"github.com/lamgor666/goboot-common/enum/JwtVerifyErrno"
	"github.com/lamgor666/goboot-common/enum/RegexConst"
	"github.com/lamgor666/goboot-common/util/castx"
	"github.com/lamgor666/goboot-common/util/jsonx"
	"github.com/lamgor666/goboot-common/util/stringx"
	"github.com/lamgor666/goboot-common/util/validatex"
	"github.com/lamgor666/goboot-dal/RateLimiter"
	"github.com/lamgor666/goboot-gin/http/error/JwtAuthError"
	"github.com/lamgor666/goboot-gin/http/error/RateLimitError"
	"github.com/lamgor666/goboot-gin/http/error/ValidateError"
	"strings"
	"time"
)

var Version = "1.0.0"
var errorHandlers = make([]ErrorHandler, 0)

func WithBuiltinErrorHandlers() {
	errorHandlers = []ErrorHandler{
		rateLimitErrorHandler{},
		jwtAuthErrorHandler{},
		validateErrorHandler{},
	}
}

func ReplaceBuiltinErrorHandler(errName string, handler ErrorHandler) {
	errName = stringx.EnsureRight(errName, "Error")
	errName = stringx.EnsureLeft(errName, "builtin.")
	handlers := make([]ErrorHandler, 0)
	var added bool

	for _, h := range errorHandlers {
		if h.GetErrorName() == errName {
			handlers = append(handlers, handler)
			added = true
			continue
		}

		handlers = append(handlers, h)
	}

	if !added {
		handlers = append(handlers, handler)
	}

	errorHandlers = handlers
}

func WithErrorHandler(handler ErrorHandler) {
	handlers := make([]ErrorHandler, 0)
	var added bool

	for _, h := range errorHandlers {
		if h.GetErrorName() == handler.GetErrorName() {
			handlers = append(handlers, handler)
			added = true
			continue
		}

		handlers = append(handlers, h)
	}

	if !added {
		handlers = append(handlers, handler)
	}

	errorHandlers = handlers
}

func WithErrorHandlers(handlers []ErrorHandler) {
	if len(handlers) < 1 {
		return
	}

	for _, handler := range handlers {
		WithErrorHandler(handler)
	}
}

func RateLimitCheck(ctx *gin.Context, handlerName string, settings interface{}) {
	var total int
	var duration time.Duration
	var limitByIp bool

	if map1, ok := settings.(map[string]interface{}); ok && len(map1) > 0 {
		total = castx.ToInt(map1["total"])

		if d1, ok := map1["duration"].(time.Duration); ok && d1 > 0 {
			duration = d1
		} else if n1, err := castx.ToInt64E(map1["duration"]); err == nil && n1 > 0 {
			duration = time.Duration(n1) * time.Millisecond
		}

		limitByIp = castx.ToBool(map1["limitByIp"])
	} else if s1, ok := settings.(string); ok && s1 != "" {
		s1 = strings.ReplaceAll(s1, "[syh]", `"`)
		map1 := jsonx.MapFrom(s1)

		if len(map1) > 0 {
			total = castx.ToInt(map1["total"])

			if d1, ok := map1["duration"].(time.Duration); ok && d1 > 0 {
				duration = d1
			} else if n1, err := castx.ToInt64E(map1["duration"]); err == nil && n1 > 0 {
				duration = time.Duration(n1) * time.Millisecond
			}

			limitByIp = castx.ToBool(map1["limitByIp"])
		}
	}

	if handlerName == "" || total < 1 || duration < 1 {
		return
	}

	id := handlerName

	if limitByIp {
		id += "@" + GetClientIp(ctx)
	}

	opts := RateLimiter.NewOptions(RatelimiterLuaFile(), RatelimiterCacheDir())
	limiter := RateLimiter.New(id, total, duration, opts)
	result := limiter.GetLimit()
	remaining := castx.ToInt(result["remaining"])

	if remaining < 0 {
		panic(RateLimitError.New(result))
	}
}

func JwtAuthCheck(ctx *gin.Context, settingsKey string) {
	if settingsKey == "" {
		return
	}

	settings := JwtSettings(settingsKey)

	if settings == nil {
		return
	}

	token := GetHeader(ctx, "Authorization")
	token = stringx.RegexReplace(token, RegexConst.SpaceSep, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		panic(JwtAuthError.New(JwtVerifyErrno.NotFound))
	}

	errno := VerifyJwt(token, settings)

	if errno < 0 {
		panic(JwtAuthError.New(errno))
	}

	return
}

func ValidateCheck(ctx *gin.Context, settings interface{}) {
	rules := make([]string, 0)
	var failfast bool

	if items, ok := settings.([]string); ok && len(items) > 0 {
		for _, s1 := range items {
			if s1 == "" || s1 == "false" {
				continue
			}

			if s1 == "true" {
				failfast = true
				continue
			}

			rules = append(rules, s1)
		}
	} else if s1, ok := settings.(string); ok && s1 != "" {
		s1 = strings.ReplaceAll(s1, "[syh]", `"`)
		entries := jsonx.ArrayFrom(s1)

		for _, entry := range entries {
			s2, ok := entry.(string)

			if !ok || s2 == "" || s2 == "false" {
				continue
			}

			if s2 == "true" {
				failfast = true
				continue
			}

			rules = append(rules, s2)
		}
	}

	if len(rules) < 1 {
		return
	}

	validator := validatex.NewValidator()
	data := GetMap(ctx)

	if failfast {
		errorTips := validatex.FailfastValidate(validator, data, rules)

		if errorTips != "" {
			panic(ValidateError.New(errorTips, true))
		}

		return
	}

	validateErrors := validatex.Validate(validator, data, rules)

	if len(validateErrors) > 0 {
		panic(ValidateError.New(validateErrors))
	}
}
