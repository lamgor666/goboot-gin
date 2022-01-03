package ImageResponse

import (
	"github.com/lamgor666/goboot-common/util/mimex"
	"io/ioutil"
)

type Impl struct {
	buf      []byte
	mimeType string
}

func FromFile(fpath string, mimeType ...string) Impl {
	buf, _ := ioutil.ReadFile(fpath)
	var _mimeType string

	if len(mimeType) > 0 {
		_mimeType = mimeType[0]
	}

	if _mimeType == "" {
		_mimeType = mimex.GetMimeType(buf)
	}

	return Impl{
		buf:      buf,
		mimeType: _mimeType,
	}
}

func FromBuffer(buf []byte, mimeType ...string) Impl {
	var _mimeType string

	if len(mimeType) > 0 {
		_mimeType = mimeType[0]
	}

	if _mimeType == "" {
		_mimeType = mimex.GetMimeType(buf)
	}

	return Impl{
		buf:      buf,
		mimeType: _mimeType,
	}
}

func (p Impl) GetContentType() string {
	return p.mimeType
}

func (p Impl) GetContents() (int, string) {
	if len(p.buf) < 1 || p.mimeType == "" {
		return 400, ""
	}

	return 200, ""
}

func (p Impl) Buffer() []byte {
	return p.buf
}
