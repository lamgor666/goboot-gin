package ValidateError

type Impl struct {
	errorTips      string
	validateErrors map[string]string
	failfast       bool
}

func New(args ...interface{}) Impl {
	var errorTips string

	for _, arg := range args {
		if s1, ok := arg.(string); ok && s1 != "" {
			errorTips = s1
			break
		}
	}

	if errorTips == "" {
		errorTips = "数据完整性验证错误"
	}

	validateErrors := map[string]string{}

	for _, arg := range args {
		if map1, ok := arg.(map[string]string); ok && len(map1) > 0 {
			validateErrors = map1
			break
		}
	}

	var failfast bool

	for _, arg := range args {
		if b1, ok := arg.(bool); ok {
			failfast = b1
			break
		}
	}

	return Impl{
		errorTips:      errorTips,
		validateErrors: validateErrors,
		failfast:       failfast,
	}
}

func (ex Impl) Error() string {
	return ex.errorTips
}

func (ex Impl) ValidateErrors() map[string]string {
	return ex.validateErrors
}

func (ex Impl) Failfast() bool {
	return ex.failfast
}
