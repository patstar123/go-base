package base

import (
	"encoding/json"
	"fmt"
)

type Result interface {
	IsOk() bool
	Code() int
	Message() string
	Error() string

	Data() any

	SetMsg(msg string) Result
	AppendMsg(msg string) Result
	AppendErr(msg string, err error) Result

	SetData(data any) Result

	IsEqual(other Result) bool
}

type result struct {
	ICode    int    `json:"code"`
	IMessage string `json:"message"`
	IData    any    `json:"data"`
}

func NewResult(code int, message string, data any) Result {
	return &result{code, message, data}
}

func UnmarshalJson(data []byte) (error, Result) {
	var res result
	err := json.Unmarshal(data, &res)
	return err, &res
}

func (r *result) IsOk() bool {
	return r != nil && r.ICode == SUCCESS.ICode
}

func (r *result) Code() int {
	if r == nil {
		return UNKNOWN.ICode
	}
	return r.ICode
}

func (r *result) Message() string {
	if r == nil {
		return UNKNOWN.IMessage
	}
	return r.IMessage
}

func (r *result) Error() string {
	if r == nil {
		return UNKNOWN.Error()
	}
	return fmt.Sprintf("(%v) %s", r.ICode, r.IMessage)
}

func (r *result) Data() any {
	if r == nil {
		return UNKNOWN.IData
	}
	return r.IData
}

func (r *result) SetMsg(msg string) Result {
	if r == nil {
		return UNKNOWN.SetMsg(msg)
	}
	return &result{r.ICode, msg, r.IData}
}

func (r *result) AppendMsg(msg string) Result {
	if r == nil {
		return UNKNOWN.AppendMsg(msg)
	}
	if r.IMessage == "" {
		return &result{r.ICode, msg, r.IData}
	} else {
		return &result{r.ICode, r.IMessage + " << " + msg, r.IData}
	}
}

func (r *result) AppendErr(msg string, err error) Result {
	if r == nil {
		return UNKNOWN.AppendErr(msg, err)
	}
	if r.IMessage == "" {
		return &result{r.ICode, msg + "(" + err.Error() + ")", r.IData}
	} else {
		return &result{r.ICode, r.IMessage + " << " + msg + "(" + err.Error() + ")", r.IData}
	}
}

func (r *result) SetData(data any) Result {
	if r == nil {
		return UNKNOWN.SetData(data)
	}
	return &result{r.ICode, r.IMessage, data}
}

func (r *result) IsEqual(other Result) bool {
	if r == nil {
		return UNKNOWN.IsEqual(other)
	}
	return r.ICode == other.Code()
}

var (
	SUCCESS = &result{0, "", nil}
	UNKNOWN = &result{-1, "未知错误", nil}

	START = &result{-2000, "start of error", nil}
	END   = &result{-3000, "end of error", nil}

	INVALID_PARAM       = &result{START.ICode - 1, "无效的参数", nil}      // invalid param
	TARGET_NOT_FOUND    = &result{START.ICode - 2, "目标不存在", nil}      // target not found
	ACTION_ILLEGAL      = &result{START.ICode - 3, "操作非法", nil}       // action illegal
	ACTION_TIMEOUT      = &result{START.ICode - 4, "操作超时", nil}       // action timeout
	ACTION_CANCELED     = &result{START.ICode - 5, "操作被取消", nil}      // action canceled
	ACTION_UNSUPPORTED  = &result{START.ICode - 6, "不支持的操作", nil}     // action unsupported
	LOGICAL_ERROR       = &result{START.ICode - 7, "逻辑错误", nil}       // logical error
	INTERNAL_ERROR      = &result{START.ICode - 10, "内部错误", nil}      // internal error
	REMOTE_SYSTEM_ERROR = &result{START.ICode - 11, "远端系统错误", nil}    // remote system error
	TRY_AGAIN_LATER     = &result{START.ICode - 12, "操作失败,稍后再试", nil} // action failed, try again later
)
