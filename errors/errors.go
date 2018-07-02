// nolint: comments
package errors

import (
	"bytes"
	native "errors"
	"fmt"
	"net/http"
)

//  RuntimeError 一个带 Code 的 error
type RuntimeError interface {
	HTTPCode() int
	Code() int
	Error() string
}

//  ErrNotFound 对象找不到
type ErrNotFound struct {
	typ string
	id  interface{}
}

func (err *ErrNotFound) HTTPCode() int {
	return http.StatusNotFound
}

func (err *ErrNotFound) Code() int {
	return http.StatusNotFound
}

func (err *ErrNotFound) Error() string {
	if nil == err.id {
		return "not found"
	}
	return "record with type is '" + err.typ + "' and id is '" + fmt.Sprint(err.id) + "' isn't found"
}

//  IsNotFound 是不是一个未找到错误
func IsNotFound(e error) bool {
	_, ok := e.(*ErrNotFound)
	if ok {
		return ok
	}
	re, ok := e.(RuntimeError)
	return ok && re.HTTPCode() == http.StatusNotFound
}

//  NotFound 创建一个 ErrNotFound
func NotFound(id interface{}, typ ...string) RuntimeError {
	if len(typ) == 0 {
		return &ErrNotFound{id: id}
	}

	return &ErrNotFound{id: id, typ: typ[0]}
}

//  New 创建一个 error
func New(msg string) error {
	return native.New(msg)
}

//  PGError postgresql error
type PGError interface {
	Error() string
	Fatal() bool
	Get(k byte) (v string)
}

//  Wrap 为 error 增加上下文信息
func Wrap(e error, s string, args ...interface{}) error {
	if "" == s {
		return e
	}
	if re, ok := e.(RuntimeError); ok {
		return NewApplicationError(re.Code(), fmt.Sprintf(s, args...)+": "+e.Error())
	}
	return native.New(fmt.Sprintf(s, args...) + ": " + e.Error())
}

//  RuntimeWrap 为 error 增加上下文信息
func RuntimeWrap(e error, s string, args ...interface{}) RuntimeError {
	if "" == s {
		return ToRuntimeError(e)
	}

	msg := fmt.Sprintf(s, args...) + ": " + e.Error()
	if re, ok := e.(RuntimeError); ok {
		return NewApplicationError(re.Code(), msg)
	}
	return NewApplicationError(http.StatusInternalServerError, msg)
}

//  MutiErrors 拼接多个错误
type MutiErrors struct {
	msg  string
	errs []error
}

func (self *MutiErrors) Error() string {
	return self.msg
}
func (self *MutiErrors) Errors() []error {
	return self.errs
}

// Concat 拼接多个错误
func Concat(errs []error, format string, args ...interface{}) error {
	if len(errs) == 1 && format == "" {
		return errs[0]
	}

	var buffer bytes.Buffer
	isFirst := true
	if format != "" {
		isFirst = false
		fmt.Fprintf(&buffer, format, args...)
	}
	for _, e := range errs {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString("\n ")
		}
		buffer.WriteString(e.Error())
	}
	return &MutiErrors{msg: buffer.String(), errs: errs}
}

// ApplicationError 应用错误
type ApplicationError struct {
	ErrCode    int    `json:"code,omitempty"`
	ErrMessage string `json:"message"`
}

func (err *ApplicationError) HTTPCode() int {
	return err.ErrCode
}

func (err *ApplicationError) Code() int {
	return err.ErrCode
}

func (err *ApplicationError) Error() string {
	return err.ErrMessage
}

// NewApplicationError 创建一个 RuntimeError
func NewApplicationError(code int, msg string) RuntimeError {
	return &ApplicationError{ErrCode: code, ErrMessage: msg}
}

// ToRuntimeError 转换成 RuntimeError
func ToRuntimeError(e error) RuntimeError {
	if re, ok := e.(RuntimeError); ok {
		return re
	}
	return &ApplicationError{ErrCode: http.StatusInternalServerError, ErrMessage: e.Error()}
}

// ToApplicationError 转换成 ApplicationError
func ToApplicationError(e error) *ApplicationError {
	if ae, ok := e.(*ApplicationError); ok {
		return ae
	}
	if re, ok := e.(RuntimeError); ok {
		return &ApplicationError{ErrCode: re.Code(), ErrMessage: re.Error()}
	}
	return &ApplicationError{ErrCode: http.StatusInternalServerError, ErrMessage: e.Error()}
}

// BadArgument 创建一个 ErrBadArgument
func BadArgument(msg string) *ApplicationError {
	return &ApplicationError{ErrCode: http.StatusBadRequest, ErrMessage: msg}
}
