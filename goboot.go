package goboot

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lamgor666/goboot-common/AppConf"
	"github.com/lamgor666/goboot-common/enum/JwtVerifyErrno"
	"github.com/lamgor666/goboot-common/enum/RegexConst"
	"github.com/lamgor666/goboot-common/util/castx"
	"github.com/lamgor666/goboot-common/util/jsonx"
	"github.com/lamgor666/goboot-common/util/numberx"
	"github.com/lamgor666/goboot-common/util/stringx"
	"github.com/lamgor666/goboot-common/util/validatex"
	"github.com/lamgor666/goboot-dal/RateLimiter"
	"github.com/lamgor666/goboot-gin/http/error/JwtAuthError"
	"github.com/lamgor666/goboot-gin/http/error/RateLimitError"
	"github.com/lamgor666/goboot-gin/http/error/ValidateError"
	"math/big"
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

func ErrorHandlers() []ErrorHandler {
	return errorHandlers
}

func LogExecuteTime(ctx *gin.Context) {
	if !ExecuteTimeLogEnabled() {
		return
	}

	elapsedTime := calcElapsedTime(ctx)

	if elapsedTime == "" {
		return
	}

	sb := strings.Builder{}
	sb.WriteString(GetMethod(ctx))
	sb.WriteString(" ")
	sb.WriteString(GetRequestUrl(ctx, true))
	sb.WriteString(", total elapsed time: " + elapsedTime)
	ExecuteTimeLogLogger().Info(sb.String())
	ctx.Header("X-Response-Time", elapsedTime)
}

func AddPoweredBy(ctx *gin.Context) {
	poweredBy := AppConf.GetString("app.poweredBy")

	if poweredBy == "" {
		return
	}

	ctx.Header("X-Powered-By", poweredBy)
}

func RateLimitCheck(ctx *gin.Context, handlerName string, settings interface{}) error {
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
		return nil
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
		return RateLimitError.New(result)
	}

	return nil
}

func JwtAuthCheck(ctx *gin.Context, settingsKey string) error {
	if settingsKey == "" {
		return nil
	}

	settings := GetJwtSettings(settingsKey)

	if settings == nil {
		return nil
	}

	token := GetHeader(ctx, "Authorization")
	token = stringx.RegexReplace(token, RegexConst.SpaceSep, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return JwtAuthError.New(JwtVerifyErrno.NotFound)
	}

	errno := VerifyJsonWebToken(token, settings)

	if errno < 0 {
		return JwtAuthError.New(errno)
	}

	return nil
}

func ValidateCheck(ctx *gin.Context, settings interface{}) error {
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
		return nil
	}

	validator := validatex.NewValidator()
	data := GetMap(ctx)

	if failfast {
		errorTips := validatex.FailfastValidate(validator, data, rules)

		if errorTips != "" {
			return ValidateError.New(errorTips, true)
		}

		return nil
	}

	validateErrors := validatex.Validate(validator, data, rules)

	if len(validateErrors) > 0 {
		return ValidateError.New(validateErrors)
	}

	return nil
}

func calcElapsedTime(ctx *gin.Context) string {
	var execStart time.Time
	v1, _ := ctx.Get("ExecStart")

	if t1, ok := v1.(time.Time); ok {
		ctx.Set("ExecStart", nil)
		execStart = t1
	}

	if execStart.IsZero() {
		return ""
	}

	n1 := big.NewFloat(time.Since(execStart).Seconds())

	if n1.Cmp(big.NewFloat(1.0)) != -1 {
		secs, _ := n1.Float64()
		return numberx.ToDecimalString(secs, 3) + "s"
	}

	n1 = n1.Mul(n1, big.NewFloat(1000.0))

	if n1.Cmp(big.NewFloat(1.0)) == -1 {
		return "0ms"
	}

	msecs, _ := n1.Float64()
	return fmt.Sprintf("%dms", castx.ToInt(msecs))
}
