package MidRecover

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lamgor666/goboot-common/AppConf"
	"github.com/lamgor666/goboot-gin"
)

func New() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			goboot.RuntimeLogger().Info("middleware run: MidRecover")
		}

		defer func() {
			r := recover()

			if r == nil {
				return
			}

			var err error

			if ex, ok := r.(error); ok {
				err = ex
			} else {
				err = fmt.Errorf("%v", r)
			}

			if err == nil {
				return
			}

			goboot.SendOutput(ctx, nil, err)
		}()

		ctx.Next()
	}
}
