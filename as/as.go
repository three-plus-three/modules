package as

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"time"
)

var ErrValueNotFoundCode = 404
var ErrValueNullCode = 461
var ErrTypeErrorCode = 460

type Error struct {
	code    int
	message string
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Error() string {
	return e.message
}

func errType(value interface{}, exceptedType string) error {
	return &Error{code: ErrTypeErrorCode, message: fmt.Sprintf("value is not a %s, actual is %T '%#v'", exceptedType, value, value)}
}

func errOverflow(value interface{}, exceptedType string) error {
	return &Error{code: ErrTypeErrorCode, message: fmt.Sprintf("value is overflow for %s, actual is '%#v'", exceptedType, value)}
}

var ErrValueNotFound = &Error{code: ErrValueNotFoundCode, message: "value isn't found"}
var ErrValueNull = &Error{code: ErrValueNullCode, message: "value is nil"}

// Map type AsSerts to `map`
func Map(value interface{}) (map[string]interface{}, error) {
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errType(value, "map")
}

func Object(value interface{}) (map[string]interface{}, error) {
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}
	if nil == value {
		return nil, ErrValueNull
	}
	return nil, errType(value, "object")
}

func Objects(v interface{}) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, 10)
	switch value := v.(type) {
	case []interface{}:
		for i, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil, errType(o, "object["+strconv.FormatInt(int64(i), 10)+"]")
			}
			results = append(results, r)
		}
	case map[string]interface{}:
		for k, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil, errType(o, "object["+k+"]")
			}
			results = append(results, r)
		}
	case []map[string]interface{}:
		return value, nil
	default:
		if nil == value {
			return nil, ErrValueNull
		}
		return nil, errType(value, "map or array")
	}
	return results, nil
}

func IsEmptyArray(value interface{}) bool {
	if a, ok := value.([]interface{}); ok {
		return len(a) == 0
	}
	if a, ok := value.([]map[string]interface{}); ok {
		return len(a) == 0
	}
	if a, ok := value.([]map[string]string); ok {
		return len(a) == 0
	}
	return true
}

func Array(value interface{}) ([]interface{}, error) {
	if a, ok := value.([]interface{}); ok {
		return a, nil
	}
	if a, ok := value.([]map[string]interface{}); ok {
		res := make([]interface{}, len(a))
		for idx, v := range a {
			res[idx] = v
		}
		return res, nil
	}
	if a, ok := value.([]map[string]string); ok {
		res := make([]interface{}, len(a))
		for idx, v := range a {
			res[idx] = v
		}
		return res, nil
	}
	return nil, errType(value, "array")
}

func Strings(value interface{}) ([]string, error) {
	if value == nil {
		return nil, ErrValueNull
	}
	switch vv := value.(type) {
	case []string:
		return vv, nil
	case []interface{}:
		results := make([]string, 0, len(vv))
		for _, v := range vv {
			s, e := String(v)
			if e != nil {
				return nil, e
			}
			results = append(results, s)
		}
		return results, nil
	}

	return nil, errType(value, "string array")
}

func IntsWithDefault(value interface{}, defValue []int) []int {
	if value == nil {
		return defValue
	}
	switch vv := value.(type) {
	case []int:
		return vv
	case []interface{}:
		results := make([]int, 0, len(vv))
		for _, v := range vv {
			s, e := Int(v)
			if e != nil {
				return defValue
			}
			results = append(results, s)
		}
		return results
	default:
		rValue := reflect.ValueOf(value)

		if rValue.Kind() == reflect.Slice {
			results := make([]int, 0, rValue.Len())
			for idx := 0; idx < rValue.Len(); idx++ {
				s, e := Int(rValue.Index(idx).Interface())
				if e != nil {
					return defValue
				}
				results = append(results, s)
			}
			return results
		}
	}

	return defValue
}

func StringsWithDefault(value interface{}, defValue []string) []string {
	if value == nil {
		return defValue
	}
	switch vv := value.(type) {
	case []string:
		return vv
	case []interface{}:
		results := make([]string, 0, len(vv))
		for _, v := range vv {
			s, e := String(v)
			if e != nil {
				return defValue
			}
			results = append(results, s)
		}
		return results
	}

	return defValue
}

func Ints(value interface{}) ([]int, error) {
	if a, ok := value.([]interface{}); ok {
		ints := make([]int, 0, len(a))
		for _, v := range a {
			i, e := Int(v)
			if nil != e {
				return nil, e
			}
			ints = append(ints, i)
		}
		return ints, nil
	}
	return nil, errType(value, "int array")
}

func Int64s(value interface{}) ([]int64, error) {
	if a, ok := value.([]interface{}); ok {
		ints := make([]int64, 0, len(a))
		for _, v := range a {
			i, e := Int64(v)
			if nil != e {
				return nil, e
			}
			ints = append(ints, i)
		}
		return ints, nil
	}
	return nil, errType(value, "int64 array")
}

func ArrayWithDefault(value interface{}, defValue []interface{}) []interface{} {
	arr, err := Array(value)
	if nil != err {
		return defValue
	}
	return arr
}

// Bool type AsSerts to `bool`
func Bool(value interface{}) (bool, error) {
	if s, ok := value.(bool); ok {
		return s, nil
	}
	if s, ok := value.(string); ok {
		switch s {
		case "TRUE", "True", "true", "YES", "Yes", "yes", "on", "enabled":
			return true, nil
		case "FALSE", "False", "false", "NO", "No", "no":
			return false, nil
		}
	}
	if nil == value {
		return false, ErrValueNull
	}
	return false, errType(value, "bool")
}

// Bool type AsSerts to `bool`
func BoolWithDefaultValue(value interface{}, defaultValue bool) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	if s, ok := value.(string); ok {
		switch s {
		case "TRUE", "True", "true", "YES", "Yes", "yes":
			return true
		case "FALSE", "False", "false", "NO", "No", "no":
			return false
		}
	}
	return defaultValue
}

func Int(value interface{}) (int, error) {
	a, err := Int32(value)
	return int(a), err
}

func Uint(value interface{}) (uint, error) {
	a, err := Uint32(value)
	return uint(a), err
}

// Int type AsSerts to `float64` then converts to `int`
func Int64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if 9223372036854775807 >= int64(v) {
			return int64(v), nil
		}
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if 9223372036854775807 >= v {
			return int64(v), nil
		}
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case []byte:
		i64, err := strconv.ParseInt(string(v), 10, 64)
		if nil == err {
			return i64, nil
		}
	case string:
		i64, err := strconv.ParseInt(v, 10, 64)
		if nil == err {
			return i64, nil
		}
	case json.Number:
		i64, err := v.Int64()
		if nil == err {
			return i64, nil
		}
		f64, err := v.Float64()
		if nil == err {
			return int64(f64), nil
		}
	}
	if nil == value {
		return 0, ErrValueNull
	}
	return 0, errType(value, "int64")
}

func Int64WithDefault(v interface{}, defaultValue int64) int64 {
	s, e := Int64(v)
	if nil != e {
		return defaultValue
	}
	return s
}

func Int32(value interface{}) (int32, error) {
	i64, err := Int64(value)
	if nil != err {
		return 0, errType(value, "int32")
	}
	if -2147483648 > i64 || 2147483647 < i64 {
		return 0, errOverflow(value, "int32")
	}
	return int32(i64), nil
}

func Int32WithDefault(value interface{}, defaultValue int32) int32 {
	i32, e := Int32(value)
	if nil == e {
		return i32
	}
	return defaultValue
}

func Int16(value interface{}) (int16, error) {
	i64, err := Int64(value)
	if nil != err {
		return 0, errType(value, "int16")
	}
	if -32768 > i64 || 32767 < i64 {
		return 0, errOverflow(value, "int16")
	}
	return int16(i64), nil
}

func Int8(value interface{}) (int8, error) {
	i64, err := Int64(value)
	if nil != err {
		return 0, errType(value, "int8")
	}
	if -128 > i64 || 127 < i64 {
		return 0, errOverflow(value, "int8")
	}
	return int8(i64), nil
}

// Uint type AsSerts to `float64` then converts to `int`
func Uint64(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case []byte:
		i64, err := strconv.ParseUint(string(v), 10, 64)
		if nil == err {
			return i64, nil
		}
	case string:
		i64, err := strconv.ParseUint(v, 10, 64)
		if nil == err {
			return i64, nil
		}
		return i64, errType(value, "uint64")

	case json.Number:
		i64, err := strconv.ParseUint(v.String(), 10, 64)
		if nil == err {
			return i64, nil
		}
		f64, err := v.Float64()
		if nil == err {
			if f64 >= 0 {
				return uint64(f64), nil
			}
			if f64 < 0 {
				if math.IsNaN(f64) {
					return 0, nil
				}
				if int64(f64) == 0 {
					return 0, nil
				}
			}
		} else {
			return 0, errType(value, "uint64")
		}
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int:
		if v >= 0 {
			return uint64(v), nil
		}
	case int8:
		if v >= 0 {
			return uint64(v), nil
		}
	case int16:
		if v >= 0 {
			return uint64(v), nil
		}
	case int32:
		if v >= 0 {
			return uint64(v), nil
		}
	case int64:
		if v >= 0 {
			return uint64(v), nil
		}
	case float32:
		if v >= 0 && 18446744073709551615 >= v {
			return uint64(v), nil
		}

		if v < 0 {
			if math.IsNaN(float64(v)) {
				return 0, nil
			}
			if int64(v) == 0 {
				return 0, nil
			}
		}
	case float64:
		if v >= 0 && 18446744073709551615 >= v {
			return uint64(v), nil
		}

		if v < 0 {
			if math.IsNaN(v) {
				return 0, nil
			}
			if int64(v) == 0 {
				return 0, nil
			}
		}
	}
	if nil == value {
		return 0, ErrValueNull
	}
	return 0, errType(value, "uint64")
}

func Uint32(value interface{}) (uint32, error) {
	ui64, err := Uint64(value)
	if nil != err {
		return 0, errType(value, "uint32")
	}
	if 4294967295 < ui64 {
		return 0, errOverflow(value, "uint32")
	}
	return uint32(ui64), nil
}

func Uint16(value interface{}) (uint16, error) {
	ui64, err := Uint64(value)
	if nil != err {
		return 0, errType(value, "uint16")
	}
	if 65535 < ui64 {
		return 0, errOverflow(value, "uint16")
	}
	return uint16(ui64), nil
}

func Uint8(value interface{}) (uint8, error) {
	ui64, err := Uint64(value)
	if nil != err {
		return 0, errType(value, "uint8")
	}
	if 255 < ui64 {
		return 0, errOverflow(value, "uint8")
	}
	return uint8(ui64), nil
}

// Uint type AsSerts to `float64` then converts to `int`
func Float64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case string:
		f64, err := strconv.ParseFloat(v, 64)
		if nil == err {
			return f64, nil
		}
		return f64, errType(value, "float64")
	case json.Number:
		return v.Float64()
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	}
	if nil == value {
		return 0, ErrValueNull
	}
	return 0, errType(value, "float64")
}

func Float32(value interface{}) (float32, error) {
	f64, err := Float64(value)
	if nil != err {
		return 0, errType(value, "float32")
	}
	return float32(f64), nil
}

// String type AsSerts to `string`
func String(value interface{}) (string, error) {
	if nil == value {
		return "", ErrValueNull
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'e', -1, 64), nil
	case float64:
		return strconv.FormatFloat(v, 'e', -1, 64), nil
	case bool:
		if v {
			return "true", nil
		} else {
			return "false", nil
		}
	}

	return "", errType(value, "string")
}

func StringWithDefault(v interface{}, defaultStr string) string {
	s, e := String(v)
	if nil != e {
		return defaultStr
	}
	return s
}

func Duration(v interface{}) (time.Duration, error) {
	if t, ok := v.(time.Duration); ok {
		return t, nil
	}

	if i, e := Int64(v); nil == e {
		return time.Duration(i), nil
	}

	s, ok := v.(string)
	if !ok {
		return 0, errType(v, "duration")
	}

	m, e := time.ParseDuration(s)
	if nil == e {
		return m, nil
	}
	return 0, errType(v, "duration")
}

func DurationWithDefault(v interface{}, defValue time.Duration) time.Duration {
	if t, ok := v.(time.Duration); ok {
		return t
	}

	if i, e := Int64(v); nil == e {
		return time.Duration(i)
	}

	s, ok := v.(string)
	if !ok {
		return defValue
	}

	m, e := time.ParseDuration(s)
	if nil == e {
		return m
	}
	return defValue
}

func Time(v interface{}) (time.Time, error) {
	if t, ok := v.(time.Time); ok {
		return t, nil
	}

	s, ok := v.(string)
	if !ok {
		return time.Time{}, errType(v, "Time")
	}

	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		m, e := time.ParseInLocation(layout, s, time.Local)
		if nil == e {
			return m, nil
		}
	}

	return time.Time{}, errType(v, "Time")
}

func TimeWithDefault(v interface{}, defValue time.Time) time.Time {
	if t, ok := v.(time.Time); ok {
		return t
	}

	s, ok := v.(string)
	if !ok {
		return defValue
	}

	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		m, e := time.ParseInLocation(layout, s, time.Local)
		if nil == e {
			return m
		}
	}

	return defValue
}

func BoolWithDefault(v interface{}, defaultValue bool) bool {
	b, e := Bool(v)
	if nil != e {
		return defaultValue
	}
	return b
}

func IntWithDefault(v interface{}, defaultValue int) int {
	i, e := Int(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func UintWithDefault(v interface{}, defaultValue uint) uint {
	u, e := Uint(v)
	if nil != e {
		return defaultValue
	}
	return u
}

func Uint32WithDefault(v interface{}, defaultValue uint32) uint32 {
	u32, e := Uint32(v)
	if nil != e {
		return defaultValue
	}
	return u32
}

func Uint64WithDefault(v interface{}, defaultValue uint64) uint64 {
	u64, e := Uint64(v)
	if nil != e {
		return defaultValue
	}
	return u64
}

func ObjectWithDefault(v interface{}, defaultValue map[string]interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return defaultValue
}

func ObjectsWithDefault(v interface{}, defaultValue []map[string]interface{}) []map[string]interface{} {
	if o, ok := v.([]map[string]interface{}); ok {
		return o
	}

	a, ok := v.([]interface{})
	if !ok {
		return defaultValue
	}

	res := make([]map[string]interface{}, 0, len(a))
	for _, value := range a {
		m, ok := value.(map[string]interface{})
		if !ok {
			return defaultValue
		}
		res = append(res, m)
	}
	return res
}

func Biginteger(value interface{}) (big.Int, error) {
	var intValue big.Int
	switch v := value.(type) {
	case string:
		_, ok := intValue.SetString(v, 10)
		if ok {
			return intValue, nil
		}
	case json.Number:
		_, ok := intValue.SetString(v.String(), 10)
		if ok {
			return intValue, nil
		}
	case int:
		intValue.SetInt64(int64(v))
		return intValue, nil
	case int64:
		intValue.SetInt64(v)
		return intValue, nil
	case uint:
		intValue.SetUint64(uint64(v))
		return intValue, nil
	case uint64:
		intValue.SetUint64(v)
		return intValue, nil
	case uint8:
		intValue.SetUint64(uint64(v))
		return intValue, nil
	case uint16:
		intValue.SetUint64(uint64(v))
		return intValue, nil
	case uint32:
		intValue.SetUint64(uint64(v))
		return intValue, nil
	case int8:
		intValue.SetInt64(int64(v))
		return intValue, nil
	case int16:
		intValue.SetInt64(int64(v))
		return intValue, nil
	case int32:
		intValue.SetInt64(int64(v))
		return intValue, nil
	case float32:
		if v >= 0 && math.MaxUint64 >= v {
			intValue.SetUint64(uint64(v))
			return intValue, nil
		} else if v < 0 && math.MinInt64 <= v {
			intValue.SetInt64(int64(v))
			return intValue, nil
		}
	case float64:
		if v >= 0 && math.MaxUint64 >= v {
			intValue.SetUint64(uint64(v))
			return intValue, nil
		} else if v < 0 && math.MinInt64 <= v {
			intValue.SetInt64(int64(v))
			return intValue, nil
		}
	}
	if nil == value {
		return intValue, ErrValueNull
	}
	return intValue, errType(value, "big.Int")
}
