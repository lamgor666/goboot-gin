package goboot

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/lamgor666/goboot-common/AppConf"
	"github.com/lamgor666/goboot-common/enum/RegexConst"
	"github.com/lamgor666/goboot-common/enum/ReqParamSecurityMode"
	"github.com/lamgor666/goboot-common/util/castx"
	"github.com/lamgor666/goboot-common/util/errorx"
	"github.com/lamgor666/goboot-common/util/jsonx"
	"github.com/lamgor666/goboot-common/util/mapx"
	"github.com/lamgor666/goboot-common/util/numberx"
	"github.com/lamgor666/goboot-common/util/slicex"
	"github.com/lamgor666/goboot-common/util/stringx"
	"github.com/lamgor666/goboot-gin/http/error/RateLimitError"
	"github.com/lamgor666/goboot-gin/http/response/AttachmentResponse"
	"github.com/lamgor666/goboot-gin/http/response/ImageResponse"
	"math"
	"math/big"
	"regexp"
	"strings"
	"time"
)

func GetMethod(ctx *gin.Context) string {
	return strings.ToUpper(ctx.Request.Method)
}

func GetHeader(ctx *gin.Context, name string) string {
	name = strings.ToLower(name)
	headers := GetHeaders(ctx)

	for headerName, headerValue := range headers {
		if strings.ToLower(headerName) == name {
			return headerValue
		}
	}

	return ""
}

func GetHeaders(ctx *gin.Context) map[string]string {
	if len(ctx.Request.Header) < 1 {
		return map[string]string{}
	}

	map1 := map[string]string{}

	for name, values := range ctx.Request.Header {
		if len(values) < 1 {
			map1[name] = ""
			continue
		}

		map1[name] = strings.Join(values, ",")
	}

	return map1
}

func GetQueryParams(ctx *gin.Context) map[string]string {
	map1 := map[string]string{}
	values := ctx.Request.URL.Query()

	if len(values) < 1 {
		return map1
	}

	for key, parts := range values {
		if key == "" {
			continue
		}

		if len(parts) < 1 {
			map1[key] = ""
		} else {
			map1[key] = parts[0]
		}
	}

	return map1
}

func GetQueryString(ctx *gin.Context, urlencode ...bool) string {
	if len(urlencode) > 0 && urlencode[0] {
		return ctx.Request.URL.RawQuery
	}

	sb := strings.Builder{}
	values := ctx.Request.URL.Query()

	for key, parts := range values {
		if key == "" {
			continue
		}

		sb.WriteString("&")
		sb.WriteString(key)
		sb.WriteString("=")

		if len(parts) > 0 {
			sb.WriteString(parts[0])
		}
	}

	qs := sb.String()

	if qs != "" {
		return qs[1:]
	}

	return ""
}

func GetRequestUrl(ctx *gin.Context, withQueryString ...bool) string {
	s1 := ctx.Request.URL.RequestURI()
	s1 = stringx.EnsureLeft(s1, "/")

	if len(withQueryString) > 0 && withQueryString[0] {
		qs := GetQueryString(ctx)

		if qs != "" {
			s1 += "?" + qs
		}
	}

	return s1
}

func GetFormData(ctx *gin.Context) map[string]string {
	map1 := map[string]string{}
	ctx.PostForm("NonExistsKey")

	if len(ctx.Request.PostForm) < 1 {
		return map1
	}

	for key, values := range ctx.Request.PostForm {
		if key == "" {
			continue
		}

		if len(values) > 0 {
			map1[key] = values[0]
		} else {
			map1[key] = ""
		}
	}

	return map1
}

func GetClientIp(ctx *gin.Context) string {
	ip := GetHeader(ctx, "X-Forwarded-For")

	if ip == "" {
		ip = GetHeader(ctx, "X-Real-IP")
	}

	if ip == "" {
		ip = ctx.ClientIP()
	}

	parts := stringx.SplitWithRegexp(strings.TrimSpace(ip), RegexConst.CommaSep)

	if len(parts) < 1 {
		return ""
	}

	return strings.TrimSpace(parts[0])
}

func Pathvariable(ctx *gin.Context, name string, defaultValue ...interface{}) string {
	var dv string

	if len(defaultValue) > 0 {
		if s1, err := castx.ToStringE(defaultValue[0]); err == nil {
			dv = s1
		}
	}

	value := ctx.Param(name)

	if value == "" {
		return dv
	}

	return value
}

func PathvariableBool(ctx *gin.Context, name string, defaultValue ...interface{}) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	if b1, err := castx.ToBoolE(ctx.Param(name)); err == nil {
		return b1
	}

	return dv
}

func PathvariableInt(ctx *gin.Context, name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToInt(ctx.Param(name), dv)
}

func PathvariableInt64(ctx *gin.Context, name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToInt64(ctx.Param(name), dv)
}

func PathvariableFloat32(ctx *gin.Context, name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToFloat32(ctx.Param(name), dv)
}

func PathvariableFloat64(ctx *gin.Context, name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	return castx.ToFloat64(ctx.Param(name), dv)
}

func ReqParam(ctx *gin.Context, name string, mode int, defaultValue ...interface{}) string {
	var dv string

	if len(defaultValue) > 0 {
		if s1, err := castx.ToStringE(defaultValue[0]); err == nil {
			dv = s1
		}
	}

	modes := []int{
		ReqParamSecurityMode.None,
		ReqParamSecurityMode.HtmlPurify,
		ReqParamSecurityMode.StripTags,
	}

	if !slicex.InIntSlice(mode, modes) {
		mode = ReqParamSecurityMode.StripTags
	}

	value := ctx.PostForm(name)

	if value == "" {
		value = ctx.Query(name)
	}

	if value == "" {
		return dv
	}

	if mode != ReqParamSecurityMode.None {
		value = stringx.StripTags(value)
	}

	return value
}

func ReqParamBool(ctx *gin.Context, name string, defaultValue ...interface{}) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	value := ctx.PostForm(name)

	if value == "" {
		value = ctx.Query(name)
	}

	if b1, err := castx.ToBoolE(value); err == nil {
		return b1
	}

	return dv
}

func ReqParamInt(ctx *gin.Context, name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	value := ctx.PostForm(name)

	if value == "" {
		value = ctx.Query(name)
	}

	return castx.ToInt(value, dv)
}

func ReqParamInt64(ctx *gin.Context, name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	value := ctx.PostForm(name)

	if value == "" {
		value = ctx.Query(name)
	}

	return castx.ToInt64(value, dv)
}

func ReqParamFloat32(ctx *gin.Context, name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	value := ctx.PostForm(name)

	if value == "" {
		value = ctx.Query(name)
	}

	return castx.ToFloat32(value, dv)
}

func ReqParamFloat64(ctx *gin.Context, name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	value := ctx.PostForm(name)

	if value == "" {
		value = ctx.Query(name)
	}

	return castx.ToFloat64(value, dv)
}

func GetJwt(ctx *gin.Context) *jwt.Token {
	token := strings.TrimSpace(GetHeader(ctx, "Authorization"))
	token = stringx.RegexReplace(token, `[\x20\t]+`, " ")

	if strings.Contains(token, " ") {
		token = stringx.SubstringAfter(token, " ")
	}

	if token == "" {
		return nil
	}

	tk, _ := ParseJwt(token)
	return tk
}

func GetRawBody(ctx *gin.Context) []byte {
	method := GetMethod(ctx)
	isPost := method == "POST"
	isPut := method == "PUT"
	isPatch := method == "PATCH"
	isDelete := method == "DELETE"
	contentType := strings.ToLower(GetHeader(ctx, "Content-Type"))
	isJson := (isPost || isPut || isPatch || isDelete) && strings.Contains(contentType, "application/json")
	isXml := (isPost || isPut || isPatch || isDelete) && (strings.Contains(contentType, "application/xmln") || strings.Contains(contentType, "text/xml"))

	if isJson || isXml {
		var buf []byte
		v1, _ := ctx.Get("requestRawBody")

		if _buf, ok := v1.([]byte); ok {
			buf = _buf
		}

		if len(buf) < 1 {
			return make([]byte, 0)
		}

		if AppConf.GetBoolean("logging.logGetRawBody") {
			RuntimeLogger().Debug("raw body: " + string(buf))
		}

		return buf
	}

	isPostForm := strings.Contains(contentType, "application/x-www-form-urlencoded")
	isMultipartForm := strings.Contains(contentType, "multipart/form-data")

	if !isPost {
		return make([]byte, 0)
	}

	if !isPostForm && !isMultipartForm {
		return make([]byte, 0)
	}

	formData := GetFormData(ctx)

	if len(formData) < 1 {
		return make([]byte, 0)
	}

	sb := strings.Builder{}

	for key, value := range formData {
		sb.WriteString("&")
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(value)
	}

	contents := sb.String()

	if contents != "" {
		contents = contents[1:]
	}

	if AppConf.GetBoolean("logging.logGetRawBody") {
		RuntimeLogger().Debug("raw body via form data: " + contents)
	}

	return []byte(contents)
}

func GetMap(ctx *gin.Context, rules ...interface{}) map[string]interface{} {
	var _rules []string

	if len(rules) > 0 {
		if a1, ok := rules[0].([]string); ok && len(a1) > 0 {
			_rules = a1
		} else if s1, ok := rules[0].(string); ok && s1 != "" {
			re := regexp.MustCompile(RegexConst.CommaSep)
			_rules = re.Split(s1, -1)
		}
	}

	method := GetMethod(ctx)
	isGet := method == "GET"
	isPost := method == "POST"
	isPut := method == "PUT"
	isPatch := method == "PATCH"
	isDelete := method == "DELETE"
	contentType := strings.ToLower(GetHeader(ctx, "Content-Type"))
	isJson := (isPost || isPut || isPatch || isDelete) && strings.Contains(contentType, "application/json")

	if isJson {
		map1 := map[string]interface{}{}
		var buf []byte
		v1, _ := ctx.Get("requestRawBody")

		if _buf, ok := v1.([]byte); ok {
			buf = _buf
		}

		if len(buf) > 0 {
			map1 = jsonx.MapFrom(buf)
		}

		if len(_rules) < 1 {
			return map1
		}

		return getMapWithRules(map1, _rules)
	}

	isXml := (isPost || isPut || isPatch || isDelete) && (strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml"))

	if isXml {
		map1 := map[string]string{}
		var buf []byte
		v1, _ := ctx.Get("requestRawBody")

		if _buf, ok := v1.([]byte); ok {
			buf = _buf
		}

		if len(buf) > 0 {
			map1 = mapx.FromXml(buf)
		}

		if len(_rules) < 1 {
			return castx.ToStringMap(map1)
		}

		return getMapWithRules(castx.ToStringMap(map1), _rules)
	}

	if isGet {
		map1 := GetQueryParams(ctx)

		if len(_rules) < 1 {
			return castx.ToStringMap(map1)
		}

		return getMapWithRules(castx.ToStringMap(map1), _rules)
	}

	if !isPost {
		return map[string]interface{}{}
	}

	isPostForm := strings.Contains(contentType, "application/x-www-form-urlencoded")
	isMultipartForm := strings.Contains(contentType, "multipart/form-data")

	if !isPostForm && !isMultipartForm {
		return map[string]interface{}{}
	}

	if len(_rules) > 0 {
		return getMapWithRules(ctx, _rules)
	}

	map1 := map[string]interface{}{}
	queryParams := GetQueryParams(ctx)

	for key, value := range queryParams {
		map1[key] = value
	}

	formData := GetFormData(ctx)

	for key, value := range formData {
		map1[key] = value
	}

	return map1
}

func DtoBind(ctx *gin.Context, dto interface{}) error {
	map1 := GetMap(ctx)

	if len(map1) < 1 {
		map1 = map[string]interface{}{"__UnknowKey__": ""}
	}

	return mapx.BindToDto(map1, dto)
}

func SendOutput(ctx *gin.Context, payload ResponsePayload, err error) {
	if err != nil {
		var handler ErrorHandler

		for _, h := range errorHandlers {
			if h.MatchError(err) {
				handler = h
				break
			}
		}

		logExecuteTime(ctx)
		addPoweredBy(ctx)

		if handler == nil {
			RuntimeLogger().Error(errorx.Stacktrace(err))
			ctx.AbortWithStatus(500)
			return
		}

		if ex, ok := err.(RateLimitError.Impl); ok {
			ex.AddSpecifyHeaders(ctx)
		}

		payload := handler.HandleError(err)
		statusCode, contents := payload.GetContents()

		if statusCode >= 400 {
			ctx.AbortWithStatus(statusCode)
			return
		}

		ctx.Render(200, render.Data{
			ContentType: payload.GetContentType(),
			Data:        []byte(contents),
		})

		return
	}

	logExecuteTime(ctx)
	addPoweredBy(ctx)

	if payload == nil {
		ctx.Render(200, render.Data{
			ContentType: "text/html; charset=utf-8",
			Data:        []byte("unsupported response payload found"),
		})

		return
	}

	statusCode, contents := payload.GetContents()

	if statusCode >= 400 {
		ctx.AbortWithStatus(statusCode)
		return
	}

	if pl, ok := payload.(AttachmentResponse.Impl); ok {
		pl.AddSpecifyHeaders(ctx)

		ctx.Render(200, render.Data{
			ContentType: pl.GetContentType(),
			Data:        pl.Buffer(),
		})

		return
	}

	if pl, ok := payload.(ImageResponse.Impl); ok {
		ctx.Render(200, render.Data{
			ContentType: pl.GetContentType(),
			Data:        pl.Buffer(),
		})

		return
	}

	ctx.Render(200, render.Data{
		ContentType: payload.GetContentType(),
		Data:        []byte(contents),
	})
}

func getMapWithRules(arg0 interface{}, rules []string) map[string]interface{} {
	var ctx *gin.Context
	var srcMap map[string]interface{}

	if _ctx, ok := arg0.(*gin.Context); ok && _ctx != nil {
		ctx = _ctx
	} else if map1, ok := arg0.(map[string]interface{}); ok && len(map1) > 0 {
		srcMap = map1
	}

	dstMap := map[string]interface{}{}
	re1 := regexp.MustCompile(`:[^:]+$`)
	re2 := regexp.MustCompile(`:[0-9]+$`)

	for _, s1 := range rules {
		typ := 1
		mode := 2
		dv := ""

		if strings.HasPrefix(s1, "i:") {
			s1 = strings.TrimPrefix(s1, "i:")
			typ = 2

			if re1.MatchString(s1) {
				dv = stringx.SubstringAfterLast(s1, ":")
				s1 = stringx.SubstringBeforeLast(s1, ":")
			}
		} else if strings.HasPrefix(s1, "d:") {
			s1 = strings.TrimPrefix(s1, "d:")
			typ = 3

			if re1.MatchString(s1) {
				dv = stringx.SubstringAfterLast(s1, ":")
				s1 = stringx.SubstringBeforeLast(s1, ":")
			}
		} else if strings.HasPrefix(s1, "s:") {
			s1 = strings.TrimPrefix(s1, "s:")

			if re2.MatchString(s1) {
				mode = castx.ToInt(stringx.SubstringAfterLast(s1, ":"), 2)
				s1 = stringx.SubstringBeforeLast(s1, ":")
			}
		} else if re2.MatchString(s1) {
			mode = castx.ToInt(stringx.SubstringAfterLast(s1, ":"), 2)
			s1 = stringx.SubstringBeforeLast(s1, ":")
		}

		if s1 == "" || strings.Contains(s1, ":") {
			continue
		}

		srcKey := s1
		dstKey := s1

		if strings.Contains(s1, "#") {
			srcKey = stringx.SubstringBefore(s1, "#")
			dstKey = stringx.SubstringAfter(s1, "#")
		}

		var paramValue interface{}

		if ctx != nil {
			paramValue = ctx.PostForm(srcKey)

			if paramValue == "" {
				paramValue = ctx.Query(srcKey)
			}
		} else if len(srcMap) > 0 {
			paramValue = srcMap[srcKey]
		}

		switch typ {
		case 1:
			value := castx.ToString(paramValue)

			if mode != 0 {
				dstMap[dstKey] = stringx.StripTags(value)
			} else {
				dstMap[dstKey] = value
			}
		case 2:
			var value int

			if n1, err := castx.ToIntE(dv); err == nil {
				value = castx.ToInt(paramValue, n1)
			} else {
				value = castx.ToInt(paramValue)
			}

			dstMap[dstKey] = value
		case 3:
			var value float64

			if n1, err := castx.ToFloat64E(dv); err == nil {
				value = castx.ToFloat64(paramValue, n1)
			} else {
				value = castx.ToFloat64(paramValue)
			}

			dstMap[dstKey] = numberx.ToDecimalString(value)
		}
	}

	return dstMap
}

func logExecuteTime(ctx *gin.Context) {
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

func addPoweredBy(ctx *gin.Context) {
	poweredBy := AppConf.GetString("app.poweredBy")

	if poweredBy == "" {
		return
	}

	ctx.Header("X-Powered-By", poweredBy)
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
