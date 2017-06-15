package errors

import (
	"bytes"
	native "errors"
	"fmt"
	"net/http"
)

type RuntimeError interface {
	Code() int
	Error() string
}

type ErrNotFound struct {
	id interface{}
}

func (err *ErrNotFound) Code() int {
	return http.StatusNotFound
}

func (err *ErrNotFound) Error() string {
	if nil == err.id {
		return "not found"
	}
	return "record with id is '" + fmt.Sprint(err.id) + "' isn't found"
}

func NotFound(id interface{}) RuntimeError {
	return &ErrNotFound{id: id}
}

func New(msg string) error {
	return native.New(msg)
}

type PGError interface {
	Error() string
	Fatal() bool
	Get(k byte) (v string)
}

func Wrap(s string, e error) error {
	if "" == s {
		return e
	}
	return native.New(s + ": " + e.Error())
}

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

func Concat(msg string, errs []error) error {
	if len(errs) == 1 {
		return errs[0]
	}
	var buffer bytes.Buffer
	isFirst := true
	if msg != "" {
		isFirst = false
		buffer.WriteString(msg)
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
