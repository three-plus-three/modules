package log

import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap/zapcore"
)

type Target interface {
	LogFields(level zapcore.Level, msg string, fields ...zapcore.Field)
}

func concatTargets(target Target, targets ...Target) Target {
	if a, ok := target.(arrayAppender); ok {
		a.appenders = append(a.appenders, targets...)
		return a
	}

	return arrayAppender{appenders: append(targets, target)}
}

type arrayAppender struct {
	appenders []Target
}

func (sl arrayAppender) LogFields(level zapcore.Level, msg string, fields ...zapcore.Field) {
	for idx := range sl.appenders {
		sl.appenders[idx].LogFields(level, msg, fields...)
	}
}

type Callback func(level zapcore.Level, msg string, fields ...zapcore.Field)

func (callback Callback) LogFields(level zapcore.Level, msg string, fields ...zapcore.Field) {
	callback(level, msg, fields...)
}

type MessagesAppender struct {
	Messages []string
}

func (sl MessagesAppender) LogFields(level zapcore.Level, msg string, fields ...zapcore.Field) {
	sl.Messages = append(sl.Messages, msg)
}

type spanAppender struct {
	span opentracing.Span
}

func (sl spanAppender) LogFields(level zapcore.Level, msg string, fields ...zapcore.Field) {
	// TODO rather than always converting the fields, we could wrap them into a lazy logger
	fa := fieldAdapter(make([]log.Field, 0, 2+len(fields)))
	fa = append(fa, log.String("event", msg))
	fa = append(fa, log.String("level", level.String()))
	for _, field := range fields {
		field.AddTo(&fa)
	}
	sl.span.LogFields(fa...)
}

type fieldAdapter []log.Field

func (fa *fieldAdapter) AddBool(key string, value bool) {
	*fa = append(*fa, log.Bool(key, value))
}

func (fa *fieldAdapter) AddFloat64(key string, value float64) {
	*fa = append(*fa, log.Float64(key, value))
}

func (fa *fieldAdapter) AddFloat32(key string, value float32) {
	*fa = append(*fa, log.Float64(key, float64(value)))
}

func (fa *fieldAdapter) AddInt(key string, value int) {
	*fa = append(*fa, log.Int(key, value))
}

func (fa *fieldAdapter) AddInt64(key string, value int64) {
	*fa = append(*fa, log.Int64(key, value))
}

func (fa *fieldAdapter) AddInt32(key string, value int32) {
	*fa = append(*fa, log.Int64(key, int64(value)))
}

func (fa *fieldAdapter) AddInt16(key string, value int16) {
	*fa = append(*fa, log.Int64(key, int64(value)))
}

func (fa *fieldAdapter) AddInt8(key string, value int8) {
	*fa = append(*fa, log.Int64(key, int64(value)))
}

func (fa *fieldAdapter) AddUint(key string, value uint) {
	*fa = append(*fa, log.Uint64(key, uint64(value)))
}

func (fa *fieldAdapter) AddUint64(key string, value uint64) {
	*fa = append(*fa, log.Uint64(key, value))
}

func (fa *fieldAdapter) AddUint32(key string, value uint32) {
	*fa = append(*fa, log.Uint64(key, uint64(value)))
}

func (fa *fieldAdapter) AddUint16(key string, value uint16) {
	*fa = append(*fa, log.Uint64(key, uint64(value)))
}

func (fa *fieldAdapter) AddUint8(key string, value uint8) {
	*fa = append(*fa, log.Uint64(key, uint64(value)))
}

func (fa *fieldAdapter) AddUintptr(key string, value uintptr)                        {}
func (fa *fieldAdapter) AddArray(key string, marshaler zapcore.ArrayMarshaler) error { return nil }
func (fa *fieldAdapter) AddComplex128(key string, value complex128)                  {}
func (fa *fieldAdapter) AddComplex64(key string, value complex64)                    {}
func (fa *fieldAdapter) AddObject(key string, value zapcore.ObjectMarshaler) error   { return nil }
func (fa *fieldAdapter) AddReflected(key string, value interface{}) error            { return nil }
func (fa *fieldAdapter) OpenNamespace(key string)                                    {}

func (fa *fieldAdapter) AddDuration(key string, value time.Duration) {
	// TODO inefficient
	*fa = append(*fa, log.String(key, value.String()))
}

func (fa *fieldAdapter) AddTime(key string, value time.Time) {
	// TODO inefficient
	*fa = append(*fa, log.String(key, value.String()))
}

func (fa *fieldAdapter) AddBinary(key string, value []byte) {
	*fa = append(*fa, log.Object(key, value))
}

func (fa *fieldAdapter) AddByteString(key string, value []byte) {
	*fa = append(*fa, log.Object(key, value))
}

func (fa *fieldAdapter) AddString(key, value string) {
	if key != "" && value != "" {
		*fa = append(*fa, log.String(key, value))
	}
}
