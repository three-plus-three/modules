package util

import "log"

func CatchError(err error) {
	if err != nil {
		log.Println("[WARN]", err)
	}
}
