package errors

import (
	"github.com/runner-mei/errors"
)

var IsPendingError = errors.IsPendingError
var IsTimeoutError = errors.IsTimeoutError
var IsNotFound = errors.IsNotFound
var IsEmptyError = errors.IsEmptyError

//// IsTimeoutError 是不是一个超时错误
//func IsTimeoutError(e error) bool {
//	if he, ok := e.(HTTPError); ok {
//		return he.HTTPCode() == ErrTimeout.HTTPCode()
//	}

//	if ex, ok := e.(RuntimeError); ok {
//		return ErrTimeout.ErrorCode() == ex.ErrorCode() ||
//			ErrTimeout.HTTPCode() == ex.ErrorCode() ||
//			ErrTimeout.HTTPCode() == ex.HTTPCode()
//	}

//	s := e.Error()
//	if pos := strings.IndexFunc(s, unicode.IsSpace); pos > 0 {
//		se := s[pos+1:]
//		return se == "time out" || se == "timeout"
//	}
//	return s == "time out" || s == "timeout"
//}

//// IsNotFound 是不是一个未找到错误
//func IsNotFound(e error) bool {
//	if e == sql.ErrNoRows {
//		return true
//	}
//	if he, ok := e.(HTTPError); ok {
//		return he.HTTPCode() == http.StatusNotFound
//	}
//	re, ok := e.(RuntimeError)
//	return ok && re.HTTPCode() == http.StatusNotFound
//}

//func IsEmptyError(e error) bool {
//	if he, ok := e.(HTTPError); ok {
//		return he.HTTPCode() == ErrResultEmpty.HTTPCode()
//	}

//	if ex, ok := e.(RuntimeError); ok {
//		return ErrCodeResultEmpty == ex.ErrorCode() ||
//			ErrCodeResultEmpty == ex.HTTPCode()
//	}
//	if e.Error() == ErrResultEmpty.Error() {
//		return true
//	}
//	return false
//}
