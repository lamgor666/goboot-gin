package AttachmentResponse

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lamgor666/goboot-common/util/mimex"
	"io/ioutil"
)

type Impl struct {
	buf                []byte
	mimeType           string
	attachmentFileName string
}

func FromFile(fpath, attachmentFileName string, mimeType ...string) Impl {
	buf, _ := ioutil.ReadFile(fpath)
	var _mimeType string

	if len(mimeType) > 0 {
		_mimeType = mimeType[0]
	}

	if _mimeType == "" {
		_mimeType = mimex.GetMimeType(buf)
	}

	return Impl{
		buf:                buf,
		mimeType:           _mimeType,
		attachmentFileName: attachmentFileName,
	}
}

func FromBuffer(buf []byte, attachmentFileName string, mimeType ...string) Impl {
	var _mimeType string

	if len(mimeType) > 0 {
		_mimeType = mimeType[0]
	}

	if _mimeType == "" {
		_mimeType = mimex.GetMimeType(buf)
	}

	return Impl{
		buf:                buf,
		mimeType:           _mimeType,
		attachmentFileName: attachmentFileName,
	}
}

func (p Impl) GetContentType() string {
	if p.mimeType == "" {
		return "application/octet-stream"
	}

	return p.mimeType
}

func (p Impl) GetContents() (int, string) {
	if len(p.buf) < 1 || p.attachmentFileName == "" {
		return 400, ""
	}

	return 200, ""
}

func (p Impl) Buffer() []byte {
	return p.buf
}

func (p Impl) AddSpecifyHeaders(ctx *gin.Context) {
	disposition := fmt.Sprintf(`attachment; filename="%s"`, p.attachmentFileName)
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(p.buf)))
	ctx.Header("Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", disposition)
	ctx.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	ctx.Header("Pragma", "public")
}
