package utils

func Get(obj any, getter func() any, failed func() any) any {
	if obj == nil {
		return failed()
	} else {
		return getter()
	}
}

func GetString(isNull func() bool, getter func() string) string {
	if isNull() {
		return ""
	} else {
		return getter()
	}
}

func GetString2(obj any, getter func() string, failed func() string) string {
	if obj == nil {
		return failed()
	} else {
		return getter()
	}
}
