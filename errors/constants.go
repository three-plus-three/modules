package errors

import (
	"net/http"
)

// 错误码
const (
	ErrCodeTimeout          = http.StatusGatewayTimeout*1000 + 1
	ErrCodeNotFound         = http.StatusNotFound * 1000
	ErrCodeInternalError    = http.StatusInternalServerError*1000 + 1
	ErrCodeDisabled         = http.StatusForbidden*1000 + 1
	ErrCodeNotAcceptable    = http.StatusNotAcceptable*1000 + 1
	ErrCodeNotImplemented   = http.StatusNotImplemented*1000 + 1
	ErrCodePending          = http.StatusResetContent*1000 + 1
	ErrCodeIsRequired       = http.StatusBadRequest*1000 + 900
	ErrCodeNetworkError     = 560000
	ErrCodeInterruptError   = 561000
	ErrCodeMultipleError    = 562000
	ErrCodeTableIsNotExists = 591000
	ErrCodeResultEmpty      = 592000
	ErrCodeKeyNotFound      = http.StatusNotFound*1000 + 501
)

func ToHttpStatus(code int) int {
	if code < 1000 {
		return code
	}
	return code / 1000
}

var (
	ErrTimeout     = NewApplicationError(ErrCodeTimeout, "time out")
	ErrResultEmpty = NewApplicationError(ErrCodeResultEmpty, "results is empty")
)