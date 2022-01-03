package HtmlResponse

type impl struct {
	contents string
}

func New(contents string) impl {
	return impl{contents: contents}
}

func (p impl) GetContentType() string {
	return "text/html; charset=utf-8"
}

func (p impl) GetContents() (int, string) {
	return 200, p.contents
}
