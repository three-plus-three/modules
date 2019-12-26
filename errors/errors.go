// nolint: comments
package errors

import (
	"database/sql"
	native "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/runner-mei/errors"
)

type ApplicationError = errors.Error
type ValidationError = errors.ValidationError
type HTTPError = errors.HTTPError

//  RuntimeError 一个带 Code 的 error
type RuntimeError interface {
	HTTPError

	ErrorCode() int
}

var _ RuntimeError = &ApplicationError{}

// NewApplicationError 创建一个 ApplicationError
func NewApplicationError(code int, msg string) *ApplicationError {
	return &ApplicationError{Code: code, Message: msg}
}

// NewRuntimeError 创建一个 RuntimeError
func NewRuntimeError(code int, msg string) RuntimeError {
	return NewApplicationError(code, msg)
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

//  NotFound 创建一个 ErrNotFound
func NotFoundWithMessage(msg string) *ApplicationError {
	if msg == "" {
		return NewApplicationError(http.StatusNotFound, "not found")
	}

	return NewApplicationError(http.StatusNotFound, msg)
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

		if aerr, ok := errs[0].(*ApplicationError); ok && aerr.HTTPCode() == errors.ErrMultipleError.HTTPCode() {
			appError = aerr
			errs = errs[1:]
		} else if aerr, ok := errs[len(errs)-1].(*ApplicationError); ok && aerr.HTTPCode() == errors.ErrMultipleError.HTTPCode() {
			appError = aerr
			errs = errs[:len(errs)-1]
		}
	}

	if appError == nil {
		appError = &ApplicationError{Code: errors.ErrMultipleError.ErrorCode(), Message: message}
	}

	for _, e := range errs {
		if me, ok := e.(*ApplicationError); ok {
			if me.HTTPCode() == errors.ErrMultipleError.HTTPCode() && me.Message == "" {
				if len(me.Internals) > 0 {
					appError.Internals = append(appError.Internals, me.Internals...)
				}
			} else {
				appError.Internals = append(appError.Internals, *me)
			}
			continue
		}

		appError.Internals = append(appError.Internals, *ToApplicationError(e, 0))
	}

	return appError
}

// ToRuntimeError 转换成 RuntimeError
func ToRuntimeError(e error, code ...int) RuntimeError {
	if re, ok := e.(RuntimeError); ok {
		return re
	}
	return toApplicationError(e, code...)
}

// ToApplicationError 转换成 ApplicationError
func ToApplicationError(e error, code ...int) *ApplicationError {
	if ae, ok := e.(*ApplicationError); ok {
		return ae
	}
	return toApplicationError(e, code...)
}

func toApplicationError(e error, code ...int) *ApplicationError {
	if he, ok := e.(interface {
		ErrorCode() int
	}); ok {
		return &ApplicationError{Cause: e, Code: he.ErrorCode(), Message: e.Error()}
	}

	if he, ok := e.(HTTPError); ok {
		return &ApplicationError{Cause: e, Code: he.HTTPCode(), Message: e.Error()}
	}

	if len(code) > 0 {
		return &ApplicationError{Cause: e, Code: code[0], Message: e.Error()}
	}
	return &ApplicationError{Cause: e, Code: http.StatusInternalServerError, Message: e.Error()}
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
		return &ApplicationError{Cause: e, Code: re.ErrorCode(), Message: msg}
	}
	if re, ok := e.(interface {
		ErrorCode() int
	}); ok {
		return &ApplicationError{Cause: e, Code: re.ErrorCode(), Message: msg}
	}
	if re, ok := e.(HTTPError); ok {
		return &ApplicationError{Cause: e, Code: re.HTTPCode(), Message: msg}
	}

	if e == sql.ErrNoRows {
		return &ApplicationError{Cause: e, Code: http.StatusNotFound, Message: msg}
	}

	return &ApplicationError{Cause: e, Code: http.StatusInternalServerError, Message: msg}
}

// BadArgument 创建一个 ErrBadArgument
func BadArgument(msg string) *ApplicationError {
	return errors.BadArgumentWithMessage(msg)
}
func BadArgumentWithMessage(msg string) *ApplicationError {
	return errors.BadArgumentWithMessage(msg)
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
		return NewApplicationError(errors.ErrInterruptError.ErrorCode(), message)
	}

	var appError *ApplicationError
	if message == "" {
		if len(errs) == 1 {
			return errs[0]
		}

		if aerr := errs[0]; aerr.HTTPCode() == errors.ErrMultipleError.HTTPCode() {
			appError = aerr
			errs = errs[1:]
		} else if aerr := errs[len(errs)-1]; aerr.HTTPCode() == errors.ErrMultipleError.HTTPCode() {
			appError = aerr
			errs = errs[:len(errs)-1]
		}
	}

	if appError == nil {
		appError = &ApplicationError{Code: errors.ErrMultipleError.ErrorCode(), Message: message}
	}

	for _, me := range errs {
		if me.HTTPCode() == errors.ErrMultipleError.HTTPCode() && me.Message == "" {
			if len(me.Internals) > 0 {
				appError.Internals = append(appError.Internals, me.Internals...)
			}
		} else {
			appError.Internals = append(appError.Internals, *me)
		}
	}

	return appError
}
