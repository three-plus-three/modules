package util

import (
	"encoding/json"
	"io"
	"log"
)

// CloseWith 捕获错误并打印
func CloseWith(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println("[WARN]", err)
	}
}

func ToJSON(a interface{}) string {
	bs, _ := json.Marshal(a)
	if len(bs) == 0 {
		return ""
	}
	return string(bs)
}
