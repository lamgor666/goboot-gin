package goboot

type ResponsePayload interface {
	GetContentType() string
	GetContents() (int, string)
}
