package MidRequestLog

import (
	"github.com/gin-gonic/gin"
	"github.com/lamgor666/goboot-common/AppConf"
	"github.com/lamgor666/goboot-gin"
	"strings"
	"time"
)

func New() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			goboot.RuntimeLogger().Info("middleware run: MidRequestLog")
		}

		ctx.Set("ExecStart", time.Now())

		if !goboot.RequestLogEnabled() {
			ctx.Next()
			return
		}

		logger := goboot.RequestLogLogger()
		sb := strings.Builder{}
		sb.WriteString(goboot.GetMethod(ctx))
		sb.WriteString(" ")
		sb.WriteString(goboot.GetRequestUrl(ctx, true))
		sb.WriteString(" from ")
		sb.WriteString(goboot.GetClientIp(ctx))
		logger.Info(sb.String())

		if goboot.LogRequestBody() {
			rawBody := goboot.GetRawBody(ctx)

			if len(rawBody) > 0 {
				logger.Debugf(string(rawBody))
			}
		}

		ctx.Next()
	}
}
