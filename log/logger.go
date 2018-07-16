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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a simplified abstraction of the zap.Logger
type Logger interface {
	Info(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)

	Infow(msg string, fields ...interface{})
	Errorw(msg string, fields ...interface{})
	Warnw(msg string, fields ...interface{})
	Fatalw(msg string, fields ...interface{})

	Infof(msg string, values ...interface{})
	Errorf(msg string, values ...interface{})
	Warnf(msg string, values ...interface{})
	Fatalf(msg string, values ...interface{})

	With(keyAndValues ...interface{}) Logger
	WithTargets(targets ...Target) Logger
	Named(name string) Logger
}

// logger delegates all calls to the underlying zap.Logger
type logger struct {
	logger  *zap.Logger
	sugared *zap.SugaredLogger
}

// Info logs an info msg with fields
func (l logger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs an error msg with fields
func (l logger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error msg with fields
func (l logger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal error msg with fields
func (l logger) Fatal(msg string, fields ...zapcore.Field) {
	l.logger.Fatal(msg, fields...)
}

// Infow logs an info msg with fields
func (l logger) Infow(msg string, fields ...interface{}) {
	l.sugared.Infow(msg, fields...)
}

// Warnw logs an error msg with fields
func (l logger) Warnw(msg string, fields ...interface{}) {
	l.sugared.Warnw(msg, fields...)
}

// Errorw logs an error msg with fields
func (l logger) Errorw(msg string, fields ...interface{}) {
	l.sugared.Errorw(msg, fields...)
}

// Fatalw logs a fatal error msg with fields
func (l logger) Fatalw(msg string, fields ...interface{}) {
	l.sugared.Fatalw(msg, fields...)
}

// Infow logs an info msg with arguments
func (l logger) Infof(msg string, args ...interface{}) {
	l.sugared.Infof(msg, args...)
}

// Warnw logs an error msg with arguments
func (l logger) Warnf(msg string, args ...interface{}) {
	l.sugared.Warnf(msg, args...)
}

// Errorw logs an error msg with arguments
func (l logger) Errorf(msg string, args ...interface{}) {
	l.sugared.Errorf(msg, args...)
}

// Fatalw logs a fatal error msg with arguments
func (l logger) Fatalf(msg string, args ...interface{}) {
	l.sugared.Fatalf(msg, args...)
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l logger) With(keyAndValues ...interface{}) Logger {
	newL := l.logger.With(sweetenFields(l.logger, keyAndValues)...)
	return logger{logger: newL, sugared: newL.Sugar()}
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l logger) WithTargets(targets ...Target) Logger {
	return appendLogger{logger: l.logger, target: arrayAppender{targets}}
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (l logger) Named(name string) Logger {
	newL := l.logger.Named(name)
	return logger{logger: newL, sugared: newL.Sugar()}
}
