package JwtAuthError

import "fmt"

type Impl struct {
	errno int
}

func New(errno int) Impl {
	return Impl{errno: errno}
}

func (ex Impl) Error() string {
	return fmt.Sprintf("jwt auth failed, errno: %d", ex.errno)
}

func (ex Impl) Errno() int {
	return ex.errno
}
