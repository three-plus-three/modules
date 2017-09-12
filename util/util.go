package util

import "log"

// CatchError 捕获错误并打印
func CatchError(err error) {
	if err != nil {
		log.Println("[WARN]", err)
	}
}
