package errors

import (
	"github.com/runner-mei/errors"
)

type ErrorBuilder = errors.ErrorBuilder

func Build(code int, msg string) *ErrorBuilder {
	return errors.Build(code, msg)
}

func ReBuildFromRuntimeError(e RuntimeError) *ErrorBuilder {
	return errors.ReBuildFromRuntimeError(e)
}

func ReBuildFromError(e error, code int) *ErrorBuilder {
	return errors.ReBuildFromError(e, code)
}

func BuildApplicationErrorFromError(e error, code int) *ApplicationError {
	return ToApplicationError(e, code)
}
