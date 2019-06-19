package echo_opentracing

import (
	"github.com/labstack/echo"
	"github.com/runner-mei/log"
)

const (
	DefaultFactoryKey = "echo-log-factory"
	DefaultLoggerKey  = "echo-log-logger"
)

type OptionFunc func(log.Logger) log.Logger

//func OutputToStrings(target *[]string) OptionFunc {
//	return OptionFunc(func(f log.Factory) log.Factory {
//		return f.OutputToStrings(target)
//	})
//}

func FactoryFromContext(c echo.Context) log.Logger {
	of := c.Get(DefaultFactoryKey)
	if of == nil {
		return nil
	}
	return of.(log.Logger)
}

func LoggerFromContext(c echo.Context, options ...OptionFunc) log.Logger {
	ol := c.Get(DefaultLoggerKey)
	if ol != nil {
		return ol.(log.Logger)
	}

	logger := FactoryFromContext(c)
	if logger == nil {
		return nil
	}

	for _, opt := range options {
		logger = opt(logger)
	}

	span := SpanFromContext(c)
	if span != nil {
		logger = log.Span(logger, span)
	}
	return logger
}

//func Factory(factory *log.Factory) echo.MiddlewareFunc {
//	return func(next echo.HandlerFunc) echo.HandlerFunc {
//		return func(c echo.Context) error {
//			c.Set(DefaultFactoryKey, factory)
//			return next(c)
//		}
//	}
//}
