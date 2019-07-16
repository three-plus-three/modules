package util

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/mitchellh/mapstructure"
)

var IsWindows = runtime.GOOS == "windows"

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

// RollbackWith 捕获错误并打印
func RollbackWith(closer interface {
	Rollback() error
}) {
	if err := closer.Rollback(); err != nil {
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
		DecodeHook: mapstructure.ComposeDecodeHookFunc(decodeHook,
			stringToTimeHookFunc(time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02 15:04:05Z07:00",
				"2006-01-02 15:04:05",
				"2006-01-02")),
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

func stringToTimeHookFunc(layouts ...string) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		s := data.(string)
		if s == "" {
			return time.Time{}, nil
		}
		for _, layout := range layouts {
			t, err := time.Parse(layout, s)
			if err == nil {
				return t, nil
			}
		}
		// Convert it by parsing
		return data, nil
	}
}

func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// FileExists 文件是否存在
func FileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// DirExists 目录是否存在
func DirExists(dir string) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		return false
	case !d.IsDir():
		return false
	}

	return true
}

type CloseFunc func() error

func (f CloseFunc) Close() error {
	if f == nil {
		return nil
	}
	return f()
}
