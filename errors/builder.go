package errors

import (
	"fmt"

	"github.com/runner-mei/errors"
)

type ErrorBuilder struct {
	code      int
	message   string
	fields    map[string]errors.ValidationError
	internals []ApplicationError
}

func (err *ErrorBuilder) WithInternalError(e error) *ErrorBuilder {
	if rerr, ok := e.(*ApplicationError); ok {
		if rerr.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) {
			err.internals = append(err.internals, rerr.Internals...)
			return err
		}
		err.internals = append(err.internals, *rerr)
		return err
	}
	err.internals = append(err.internals, *ToApplicationError(e))
	return err
}

func (err *ErrorBuilder) WithInternalErrors(internals []*ApplicationError) *ErrorBuilder {
	if len(internals) > 0 {
		for _, e := range internals {
			err.internals = append(err.internals, *e)
		}
	}
	return err
}

func (err *ErrorBuilder) WithField(nm string, v interface{}) *ErrorBuilder {
	if nil == err.fields {
		err.fields = map[string]errors.ValidationError{}
	}
	err.fields[nm] = errors.ValidationError{Message: fmt.Sprint(v)}
	return err
}

func (err *ErrorBuilder) Fields() map[string]errors.ValidationError {
	return err.fields
}

func (err *ErrorBuilder) FieldsWithDefault() map[string]errors.ValidationError {
	if nil == err.fields {
		err.fields = map[string]errors.ValidationError{}
	}
	return err.fields
}

func (err *ErrorBuilder) Build() *ApplicationError {
	var fields map[string]errors.ValidationError
	var internals []ApplicationError
	if len(err.fields) > 0 {
		fields = err.fields
	}

	if len(err.internals) > 0 {
		internals = err.internals
	}

	return &ApplicationError{
		Code:      err.code,
		Message:   err.message,
		Fields:    fields,
		Internals: internals,
	}
}

func Build(code int, msg string) *ErrorBuilder {
	return &ErrorBuilder{
		code:    code,
		message: msg,
	}
}

func ReBuildFromRuntimeError(e RuntimeError) *ErrorBuilder {
	var fields map[string]errors.ValidationError
	var internals []ApplicationError
	if err, ok := e.(*ApplicationError); ok {
		if len(err.Fields) > 0 {
			fields = map[string]errors.ValidationError{}
			for k, v := range err.Fields {
				fields[k] = v
			}
		}

		if len(err.Internals) > 0 {
			internals = make([]ApplicationError, len(err.Internals))
			copy(internals, err.Internals)
		}
	}
	return &ErrorBuilder{
		code:      e.ErrorCode(),
		message:   e.Error(),
		fields:    fields,
		internals: internals,
	}
}

func ReBuildFromError(e error, code int) *ErrorBuilder {
	if err, ok := e.(RuntimeError); ok {
		return ReBuildFromRuntimeError(err)
	}
	return &ErrorBuilder{
		code:    code,
		message: e.Error(),
	}
}

func BuildApplicationErrorFromError(e error, code int) *ApplicationError {
	return ToApplicationError(e, code)
}
