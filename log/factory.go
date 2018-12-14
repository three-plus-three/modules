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
	"context"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Factory is the default logging wrapper that can create
// logger instances either for a given Context or context-less.
type Factory struct {
	appenders []Target
	logger    *zap.Logger
}

// NewFactory creates a new Factory.
func NewFactory(logger *zap.Logger) Factory {
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	return Factory{logger: logger}
}

// Bg creates a context-unaware logger.
func (b Factory) New() Logger {
	if len(b.appenders) == 0 {
		return logger{logger: b.logger, sugared: b.logger.Sugar()}
	}
	return appendLogger{target: arrayAppender{b.appenders}, logger: b.logger}
}

// For returns a context-aware Logger. If the context
// contains an OpenTracing span, all logging calls are also
// echo-ed into the span.
func (b Factory) For(ctx context.Context) Factory {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		return b.Span(span)
	}
	return b
}

// Span returns a span Logger, all logging calls are also
// echo-ed into the span.
func (b Factory) Span(span opentracing.Span) Factory {
	if span != nil {
		return Factory{logger: b.logger, appenders: append(b.appenders, spanAppender{span})}
	}
	return b
}

// Span returns a span Logger, all logging calls are also
// echo-ed into the span.
func (b Factory) OutputToStrings(target *[]string) Factory {
	return Factory{logger: b.logger, appenders: append(b.appenders, Callback(func(level zapcore.Level, msg string, fields ...zapcore.Field) {
		switch level {
		case zapcore.WarnLevel:
			msg = "警告：" + msg
		case zapcore.ErrorLevel:
			msg = "错误：" + msg
		}
		*target = append(*target, msg)
	}))}
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (b Factory) With(keyAndValues ...interface{}) Factory {
	return Factory{logger: b.logger.With(sweetenFields(b.logger, keyAndValues)...), appenders: b.appenders}
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (b Factory) Named(name string) Factory {
	return Factory{logger: b.logger.Named(name), appenders: b.appenders}
}

type contextKey struct{}

var activeFactoryKey = contextKey{}

// ContextWithFactory returns a new `context.Context` that holds a reference to
// `Factory`'s FactoryContext.
func ContextWithFactory(ctx context.Context, factory *Factory) context.Context {
	return context.WithValue(ctx, activeFactoryKey, factory)
}

// FactoryFromContext returns the `Factory` previously associated with `ctx`, or
// `nil` if no such `Factory` could be found.
//
// NOTE: context.Context != SpanContext: the former is Go's intra-process
// context propagation mechanism, and the latter houses OpenTracing's per-Factory
// identity and baggage information.
func FactoryFromContext(ctx context.Context) *Factory {
	val := ctx.Value(activeFactoryKey)
	if sp, ok := val.(*Factory); ok {
		return sp
	}
	return nil
}
