package util

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"reflect"

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

func decodeHook(from reflect.Kind, to reflect.Kind, v interface{}) (interface{}, error) {
	if from == reflect.String && to == reflect.Bool {
		return v.(string) == "on", nil
	}
	return v, nil
}

func ToStruct(rawVal interface{}, row map[string]interface{}) (err error) {
	config := &mapstructure.DecoderConfig{
		DecodeHook:       decodeHook,
		Metadata:         nil,
		Result:           rawVal,
		TagName:          "json",
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(row)
}
