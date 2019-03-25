// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"fmt"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type appendLogger struct {
	logger Logger
	target Target
}

// Panic logs an panic msg with fields and panic
func (l appendLogger) Panic(msg string, fields ...zapcore.Field) {
	l.target.LogFields(zap.PanicLevel, msg, fields...)
	l.logger.Panic(msg, fields...)
}

func (l appendLogger) Debug(msg string, fields ...zapcore.Field) {
	l.target.LogFields(zap.DebugLevel, msg, fields...)
	l.logger.Debug(msg, fields...)
}

func (l appendLogger) Info(msg string, fields ...zapcore.Field) {
	l.target.LogFields(zap.InfoLevel, msg, fields...)
	l.logger.Info(msg, fields...)
}

func (l appendLogger) Warn(msg string, fields ...zapcore.Field) {
	l.target.LogFields(zap.WarnLevel, msg, fields...)
	l.logger.Warn(msg, fields...)
}

func (l appendLogger) Error(msg string, fields ...zapcore.Field) {
	l.target.LogFields(zap.ErrorLevel, msg, fields...)
	l.logger.Error(msg, fields...)
}

func (l appendLogger) Fatal(msg string, fields ...zapcore.Field) {
	l.target.LogFields(zap.FatalLevel, msg, fields...)
	l.logger.Fatal(msg, fields...)
}

// Debugw logs an debug msg with fields
func (l appendLogger) Debugw(msg string, keyAndValues ...interface{}) {
	fields := sweetenFields(l, keyAndValues)
	l.Debug(msg, fields...)
}

// Infow logs an info msg with fields
func (l appendLogger) Infow(msg string, keyAndValues ...interface{}) {
	fields := sweetenFields(l, keyAndValues)
	l.Info(msg, fields...)
}

// Warnw logs an error msg with fields
func (l appendLogger) Warnw(msg string, keyAndValues ...interface{}) {
	fields := sweetenFields(l, keyAndValues)
	l.Warn(msg, fields...)
}

// Errorw logs an error msg with fields
func (l appendLogger) Errorw(msg string, keyAndValues ...interface{}) {
	fields := sweetenFields(l, keyAndValues)
	l.Error(msg, fields...)
}

// Fatalw logs a fatal error msg with fields
func (l appendLogger) Fatalw(msg string, keyAndValues ...interface{}) {
	fields := sweetenFields(l, keyAndValues)
	l.Fatal(msg, fields...)
}

// Debugf logs an info msg with fields
func (l appendLogger) Debugf(msg string, values ...interface{}) {
	l.Debug(fmt.Sprintf(msg, values))
}

// Infof logs an info msg with fields
func (l appendLogger) Infof(msg string, values ...interface{}) {
	l.Info(fmt.Sprintf(msg, values))
}

// Warnf logs an error msg with fields
func (l appendLogger) Warnf(msg string, values ...interface{}) {
	l.Warn(fmt.Sprintf(msg, values))
}

// Errorf logs an error msg with fields
func (l appendLogger) Errorf(msg string, values ...interface{}) {
	l.Error(fmt.Sprintf(msg, values))
}

// Fatalf logs a fatal error msg with fields
func (l appendLogger) Fatalf(msg string, values ...interface{}) {
	l.Fatal(fmt.Sprintf(msg, values))
}

// // With creates a child logger, and optionally adds some context fields to that logger.
// func (l appendLogger) With(fields ...zapcore.Field) Logger {
// 	return appendLogger{logger: l.logger.With(fields...), target: l.target}
// }

// With creates a child logger, and optionally adds some context fields to that logger.
func (l appendLogger) With(keyAndValues ...interface{}) Logger {
	return appendLogger{logger: l.logger.With(keyAndValues...), target: l.target}
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l appendLogger) WithTargets(targets ...Target) Logger {
	if len(targets) > 0 {
		return l
	}
	return appendLogger{logger: l.logger, target: ConcatTargets(l.target, targets...)}
}

func (l appendLogger) Named(name string) Logger {
	return appendLogger{logger: l.logger.Named(name), target: l.target}
}

const (
	_oddNumberErrMsg    = "Ignored key without a value."
	_nonStringKeyErrMsg = "Ignored key-value pairs with non-string keys."
)

func sweetenFields(base Logger, args []interface{}) []zapcore.Field {
	if len(args) == 0 {
		return nil
	}

	// Allocate enough space for the worst case; if users pass only structured
	// fields, we shouldn't penalize them with extra allocations.
	fields := make([]zapcore.Field, 0, len(args))
	var invalid invalidPairs

	for i := 0; i < len(args); {
		// This is a strongly-typed field. Consume it and move on.
		if f, ok := args[i].(zapcore.Field); ok {
			fields = append(fields, f)
			i++
			continue
		}

		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			base.Panic(_oddNumberErrMsg, Any("ignored", args[i]))
			break
		}

		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			// Subsequent errors are likely, so allocate once up front.
			if cap(invalid) == 0 {
				invalid = make(invalidPairs, 0, len(args)/2)
			}
			invalid = append(invalid, invalidPair{i, key, val})
		} else {
			fields = append(fields, Any(keyStr, val))
		}
		i += 2
	}

	// If we encountered any invalid key-value pairs, log an error.
	if len(invalid) > 0 {
		base.Panic(_nonStringKeyErrMsg, zap.Array("invalid", invalid))
	}
	return fields
}

type invalidPair struct {
	position   int
	key, value interface{}
}

func (p invalidPair) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("position", int64(p.position))
	zap.Any("key", p.key).AddTo(enc)
	zap.Any("value", p.value).AddTo(enc)
	return nil
}

type invalidPairs []invalidPair

func (ps invalidPairs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	var err error
	for i := range ps {
		err = multierr.Append(err, enc.AppendObject(ps[i]))
	}
	return err
}
