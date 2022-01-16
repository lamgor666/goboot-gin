package MidCors

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/lamgor666/goboot-common/util/castx"
	"github.com/lamgor666/goboot-common/util/slicex"
	"github.com/lamgor666/goboot-gin"
	"strconv"
	"strings"
)

func New() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		settings := goboot.CorsSettings()

		if settings == nil {
			ctx.Next()
			return
		}

		var allowedOrigins string

		if slicex.InStringSlice("*", settings.AllowedOrigins()) {
			allowedOrigins = "*"
		} else if len(settings.AllowedOrigins()) > 0 {
			allowedOrigins = strings.Join(settings.AllowedOrigins(), ", ")
		}

		if allowedOrigins != "" {
			ctx.Header("Access-Control-Allow-Origin", allowedOrigins)
		}

		if len(settings.AllowedMethods()) > 0 {
			ctx.Header("Access-Control-Allow-Methods", strings.Join(settings.AllowedMethods(), ", "))
		}

		if len(settings.AllowedHeaders()) > 0 {
			ctx.Header("Access-Control-Allow-Headers", strings.Join(settings.AllowedHeaders(), ", "))
		}

		if len(settings.ExposedHeaders()) > 0 {
			ctx.Header("Access-Control-Expose-Headers", strings.Join(settings.ExposedHeaders(), ", "))
		}

		if settings.AllowCredentials() {
			ctx.Header("Access-Control-Allow-Credentials", "true")
		}

		if settings.MaxAge() > 0 {
			n1 := castx.ToInt(settings.MaxAge().Seconds())
			ctx.Header("Access-Control-Max-Age", strconv.Itoa(n1))
		}

		if strings.ToUpper(ctx.Request.Method) == "OPTIONS" {
			ctx.Render(200, render.Data{
				ContentType: "application/json; charset=utf-8",
				Data:        []byte(`{"code":200}`),
			})
		} else {
			ctx.Next()
		}
	}
}
