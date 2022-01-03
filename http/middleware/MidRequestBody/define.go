package MidRequestBody

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/lamgor666/goboot-gin"
	"io/ioutil"
	"strings"
)

func New() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := goboot.GetMethod(ctx)
		isPost := method == "POST"
		isPut := method == "PUT"
		isPatch := method == "PATCH"
		isDelete := method == "DELETE"
		contentType := strings.ToLower(goboot.GetHeader(ctx, "Content-Type"))
		isJson := (isPost || isPut || isPatch || isDelete) && strings.Contains(contentType, "application/json")
		isXml := (isPost || isPut || isPatch || isDelete) && (strings.Contains(contentType, "application/xmln") || strings.Contains(contentType, "text/xml"))

		if isJson || isXml {
			if buf, err := ctx.GetRawData(); err == nil && len(buf) > 0 {
				ctx.Set("requestRawBody", buf)
				ctx.Request.Body = ioutil.NopCloser(bytes.NewReader(buf))
			}

			return
		}

		ctx.Next()
	}
}
