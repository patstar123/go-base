package utils

import "runtime"

func CurrentFuncName() string {
	return CallerName(1)
}

func ParentFuncName() string {
	return CallerName(2)
}

func CallerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}
	return f.Name()
}
