package ecode

import (
	"encoding/json"
	"fmt"
)

type Response struct {
	//返回自定义code
	Code int `json:"code,omitempty"`
	//http Code
	HttpCode int `json:"-"`
	//错误消息
	Message string `json:"message,omitempty"`
	//请求唯一id
	Identifier string `json:"-"`
}

func New(code, httpCode int, Message string, identifier string) *Response {
	e := &Response{Code: code, HttpCode: httpCode, Message: Message, Identifier: identifier}
	return e
}

func (e *Response) Error() string {
	return fmt.Sprintf("code: %d, Message: %s, httpCode:%d, identifier:%s", e.Code, e.Message, e.HttpCode, e.Identifier)
}

func (e *Response) Data() string {
	data, _ := json.Marshal(e)
	return string(data)
}

func (e *Response) SetMessage(Message string) error {
	return &Response{
		Code:       e.Code,
		HttpCode:   e.HttpCode,
		Message:    Message,
		Identifier: e.Identifier,
	}
}

func (e *Response) WithMessage(Message string) error {
	return &Response{
		Code:       e.Code,
		HttpCode:   e.HttpCode,
		Message:    fmt.Sprintf("%s %s", e.Message, Message),
		Identifier: e.Identifier,
	}
}

func (e *Response) error() string {
	return fmt.Sprintf("code: %d, Message: %s", e.Code, e.Message)
}
