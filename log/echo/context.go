package echo_log

import (
	"github.com/labstack/echo"
	"github.com/three-plus-three/modules/log"
	echo_opentracing "github.com/three-plus-three/modules/opentracing/echo"
)

const (
	DefaultFactoryKey = "echo-log-factory"
	DefaultLoggerKey  = "echo-log-logger"
)

type OptionFunc func(log.Factory) log.Factory

func OutputToStrings(target *[]string) OptionFunc {
	return OptionFunc(func(f log.Factory) log.Factory {
		return f.OutputToStrings(target)
	})
}

func FactoryFromContext(c echo.Context) *log.Factory {
	of := c.Get(DefaultFactoryKey)
	if of == nil {
		return nil
	}
	return of.(*log.Factory)
}

func LoggerFromContext(c echo.Context, options ...OptionFunc) log.Logger {
	ol := c.Get(DefaultLoggerKey)
	if ol != nil {
		return ol.(log.Logger)
	}

	ofactory := FactoryFromContext(c)
	if ofactory == nil {
		return nil
	}

	factory := *ofactory
	for _, opt := range options {
		factory = opt(factory)
	}

	span := echo_opentracing.SpanFromContext(c)
	if span != nil {
		factory = factory.Span(span)
	}
	return factory.New()
}

func Factory(factory *log.Factory) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(DefaultFactoryKey, factory)
			return next(c)
		}
	}
}
