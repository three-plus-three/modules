package errors

import (
	"net/http"
	"strings"
	"unicode"
)

// IsTimeoutError 是不是一个超时错误
func IsTimeoutError(e error) bool {
	if ex, ok := e.(RuntimeError); ok {
		return ErrTimeout.Code() == ex.Code() ||
			ErrTimeout.HTTPCode() == ex.Code() ||
			ErrTimeout.HTTPCode() == ex.Code()/1000
	}

	s := e.Error()
	if pos := strings.IndexFunc(s, unicode.IsSpace); pos > 0 {
		se := s[pos+1:]
		return se == "time out" || se == "timeout"
	}
	return s == "time out" || s == "timeout"
}

// IsNotFound 是不是一个未找到错误
func IsNotFound(e error) bool {
	re, ok := e.(RuntimeError)
	return ok && re.HTTPCode() == http.StatusNotFound
}

func IsEmptyError(e error) bool {
	if ex, ok := e.(RuntimeError); ok {
		return ErrCodeResultEmpty == ex.Code() ||
			ErrCodeResultEmpty == ex.Code()/1000
	}
	if e.Error() == ErrResultEmpty.Error() {
		return true
	}
	return false
}
