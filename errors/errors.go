// nolint: comments
package errors

import (
	"database/sql"
	native "errors"
	"fmt"
	"net/http"
	"strings"
)

type HTTPError interface {
	HTTPCode() int
}

//  RuntimeError 一个带 Code 的 error
type RuntimeError interface {
	HTTPError
	error

	Code() int
}

//  NotFound 创建一个 ErrNotFound
func NotFound(id interface{}, typ ...string) *ApplicationError {
	if len(typ) == 0 {
		if id == nil {
			return NewApplicationError(http.StatusNotFound, "not found")
		}

		return NewApplicationError(http.StatusNotFound, "record with id is '"+fmt.Sprint(id)+"' isn't found")
	}

	return NewApplicationError(http.StatusNotFound, "record with type is '"+typ[0]+"' and id is '"+fmt.Sprint(id)+"' isn't found")
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
	return RuntimeWrap(e, s, args...)
}

//  RuntimeWrap 为 error 增加上下文信息
func RuntimeWrap(e error, s string, args ...interface{}) RuntimeError {
	if "" == s {
		return ToRuntimeError(e)
	}

	msg := fmt.Sprintf(s, args...) + ": " + e.Error()
	if re, ok := e.(RuntimeError); ok {
		return &ApplicationError{Parent: e, ErrCode: re.Code(), ErrMessage: msg}
	}

	if e == sql.ErrNoRows {
		return &ApplicationError{Parent: e, ErrCode: http.StatusNotFound, ErrMessage: msg}
	}

	return &ApplicationError{Parent: e, ErrCode: http.StatusInternalServerError, ErrMessage: msg}
}

// Concat 拼接多个错误
func Concat(errs ...error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return ErrArray(errs)
	}
}

func ErrAppend(errs []error, format string, args ...interface{}) error {
	return ErrArray(errs, fmt.Sprintf(format, args...))
}

func ErrArray(errs []error, errMessage ...string) error {
	var message string
	if len(errMessage) > 0 {
		message = strings.Join(errMessage, " ")
	}

	if len(errs) == 0 {
		if message == "" {
			panic("Concat Fail")
		}
		return New(message)
	}

	var appError *ApplicationError
	if message == "" {
		if len(errs) == 1 {
			return errs[0]
		}

		if aerr, ok := errs[0].(*ApplicationError); ok && aerr.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) {
			appError = aerr
			errs = errs[1:]
		} else if aerr, ok := errs[len(errs)-1].(*ApplicationError); ok && aerr.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) {
			appError = aerr
			errs = errs[:len(errs)-1]
		}
	}

	if appError == nil {
		appError = &ApplicationError{ErrCode: ErrCodeMultipleError, ErrMessage: message}
	}

	for _, e := range errs {
		if me, ok := e.(*ApplicationError); ok {
			if me.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) && me.ErrMessage == "" {
				if len(me.Internals) > 0 {
					appError.Internals = append(appError.Internals, me.Internals...)
				}
			} else {
				appError.Internals = append(appError.Internals, me)
			}

			continue
		}

		appError.Internals = append(appError.Internals, ToApplicationError(e, 0))
	}

	return appError
}

// ApplicationError 应用错误
type ApplicationError struct {
	Parent     error                  `json:"-"`
	ErrCode    int                    `json:"code,omitempty"`
	ErrMessage string                 `json:"message"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Internals  []*ApplicationError    `json:"internals,omitempty"`
}

func (err *ApplicationError) HTTPCode() int {
	if err.ErrCode > 1000 {
		return err.ErrCode / 1000
	}
	return err.ErrCode
}

func (err *ApplicationError) Code() int {
	return err.ErrCode
}

func (err *ApplicationError) Error() string {
	if err.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) {
		var buffer strings.Builder
		if err.ErrMessage != "" {
			buffer.WriteString(err.ErrMessage)
		} else {
			buffer.WriteString("muti error:")
		}
		for _, e := range err.Internals {
			buffer.WriteString("\r\n  ")
			buffer.WriteString(e.Error())
		}
		return buffer.String()
	}
	return err.ErrMessage
}

// NewRuntimeError 创建一个 RuntimeError
func NewRuntimeError(code int, msg string) RuntimeError {
	return &ApplicationError{ErrCode: code, ErrMessage: msg}
}

// NewApplicationError 创建一个 ApplicationError
func NewApplicationError(code int, msg string) *ApplicationError {
	return &ApplicationError{ErrCode: code, ErrMessage: msg}
}

// ToRuntimeError 转换成 RuntimeError
func ToRuntimeError(e error, code ...int) RuntimeError {
	if re, ok := e.(RuntimeError); ok {
		return re
	}
	if he, ok := e.(interface {
		Code() int
	}); ok {
		return &ApplicationError{Parent: e, ErrCode: he.Code(), ErrMessage: e.Error()}
	}
	if he, ok := e.(interface {
		HTTPCode() int
	}); ok {
		return &ApplicationError{Parent: e, ErrCode: he.HTTPCode(), ErrMessage: e.Error()}
	}
	if len(code) > 0 {
		return &ApplicationError{Parent: e, ErrCode: code[0], ErrMessage: e.Error()}
	}
	return &ApplicationError{Parent: e, ErrCode: http.StatusInternalServerError, ErrMessage: e.Error()}
}

// ToApplicationError 转换成 ApplicationError
func ToApplicationError(e error, code ...int) *ApplicationError {
	if ae, ok := e.(*ApplicationError); ok {
		return ae
	}
	if he, ok := e.(interface {
		Code() int
	}); ok {
		return &ApplicationError{Parent: e, ErrCode: he.Code(), ErrMessage: e.Error()}
	}

	if he, ok := e.(interface {
		HTTPCode() int
	}); ok {
		return &ApplicationError{Parent: e, ErrCode: he.HTTPCode(), ErrMessage: e.Error()}
	}

	if len(code) > 0 {
		return &ApplicationError{Parent: e, ErrCode: code[0], ErrMessage: e.Error()}
	}
	return &ApplicationError{Parent: e, ErrCode: http.StatusInternalServerError, ErrMessage: e.Error()}
}

// BadArgument 创建一个 ErrBadArgument
func BadArgument(msg string) *ApplicationError {
	return &ApplicationError{ErrCode: http.StatusBadRequest, ErrMessage: msg}
}

func Is(realError, exceptedError error) bool {
	if realError == exceptedError {
		return true
	}

	ae, ok := realError.(*ApplicationError)
	if !ok || ae.Parent == nil {
		return false
	}

	return Is(ae.Parent, exceptedError)
}

func ConcatApplicationErrors(errs []*ApplicationError, errMessage ...string) *ApplicationError {
	var message string
	if len(errMessage) > 0 {
		message = strings.Join(errMessage, " ")
	}

	if len(errs) == 0 {
		if message == "" {
			panic("Concat Fail")
		}
		return NewApplicationError(ErrCodeInternalError, message)
	}

	var appError *ApplicationError
	if message == "" {
		if len(errs) == 1 {
			return errs[0]
		}

		if aerr := errs[0]; aerr.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) {
			appError = aerr
			errs = errs[1:]
		} else if aerr := errs[len(errs)-1]; aerr.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) {
			appError = aerr
			errs = errs[:len(errs)-1]
		}
	}

	if appError == nil {
		appError = &ApplicationError{ErrCode: ErrCodeMultipleError, ErrMessage: message}
	}

	for _, me := range errs {
		if me.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) && me.ErrMessage == "" {
			if len(me.Internals) > 0 {
				appError.Internals = append(appError.Internals, me.Internals...)
			}
		} else {
			appError.Internals = append(appError.Internals, me)
		}
	}

	return appError
}
