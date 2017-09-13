package util

import (
	"io"
	"log"
)

// CloseWith 捕获错误并打印
func CloseWith(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println("[WARN]", err)
	}
}
