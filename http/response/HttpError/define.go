package HttpError

type impl struct {
	statusCode int
}

func New(statusCode int) impl {
	return impl{statusCode: statusCode}
}

func (p impl) GetContentType() string {
	return ""
}

func (p impl) GetContents() (int, string) {
	return p.statusCode, ""
}
