package util

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"

	"github.com/mitchellh/mapstructure"
)

// CloseWith 捕获错误并打印
func CloseWith(closer io.Closer) {
	if err := closer.Close(); err != nil {
		if err == sql.ErrTxDone {
			return
		}

		log.Println("[WARN]", err)
		panic(err)
	}
}

func ToJSON(a interface{}) string {
	bs, _ := json.Marshal(a)
	if len(bs) == 0 {
		return ""
	}
	return string(bs)
}

func ToStruct(rawVal interface{}, row map[string]interface{}) (err error) {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   rawVal,
		TagName:  "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(row)
}
