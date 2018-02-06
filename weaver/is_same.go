package weaver

import "errors"

// ErrAlreadyClosed  server is closed
var ErrAlreadyClosed = errors.New("server is closed")

func isSame(allItems, items ValueType) bool {
	return false
}
