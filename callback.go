package base

type Callback func(result Result)

func (c Callback) On(result Result) {
	if c != nil {
		c(result)
	}
}

func (c Callback) OnSuccess() {
	if c != nil {
		c(SUCCESS)
	}
}

func (c Callback) OnSuccessD(data any) {
	if c != nil {
		c(SUCCESS.SetData(data))
	}
}
