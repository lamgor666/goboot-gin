package JsonResponse

import (
	"github.com/lamgor666/goboot-common/util/jsonx"
	"strings"
)

type impl struct {
	payload interface{}
}

func New(payload interface{}) impl {
	return impl{payload: payload}
}

func (p impl) GetContentType() string {
	return "application/json; charset=utf-8"
}

func (p impl) GetContents() (statusCode int, contents string) {
	statusCode = 200

	if s1, ok := p.payload.(string); ok {
		s1 = strings.TrimSpace(s1)

		if p.isJson(s1) {
			contents = s1
			return
		}
	}

	opts := jsonx.NewToJsonOption().HandleTimeField().StripZeroTimePart()
	contents = strings.TrimSpace(jsonx.ToJson(p.payload, opts))

	if !p.isJson(contents) {
		contents = "{}"
	}

	return
}

func (p impl) isJson(contents string) bool {
	var flag bool

	if strings.HasPrefix(contents, "{") && strings.HasSuffix(contents, "}") {
		flag = true
	} else if strings.HasPrefix(contents, "[") && strings.HasSuffix(contents, "]") {
		flag = true
	}

	return flag
}
