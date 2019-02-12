package echo_opentracing

import (
	"github.com/labstack/echo"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	//jaegercfg "github.com/uber/jaeger-client-go/config"
	//"github.com/uber/jaeger-lib/metrics"
)

const DefaultKey = "echo-opentracing-span"

// func InitGlobalTracer(serviceName, addr string) io.Closer {
// 	// Sample configuration for testing. Use constant sampling to sample every trace
// 	// and enable LogSpan to log every span via configured Logger.
// 	cfg := jaegercfg.Configuration{
// 		Sampler: &jaegercfg.SamplerConfig{
// 			Type:  jaeger.SamplerTypeConst,
// 			Param: 1,
// 		},
// 		Reporter: &jaegercfg.ReporterConfig{
// 			LogSpans: true,
// 		},
// 	}

// 	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
// 	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
// 	// frameworks.
// 	jLogger := &jaegerLogger{}
// 	jMetricsFactory := metrics.NullFactory

// 	metricsFactory := metrics.NewLocalFactory(0)
// 	metrics := jaeger.NewMetrics(metricsFactory, nil)

// 	sender, err := jaeger.NewUDPTransport(addr, 0)
// 	if err != nil {
// 		log.Printf("could not initialize jaeger sender: %s", err.Error())
// 		return nil
// 	}

// 	repoter := jaeger.NewRemoteReporter(sender, jaeger.ReporterOptions.Metrics(metrics))

// 	// Initialize tracer with a logger and a metrics factory
// 	closer, err := cfg.InitGlobalTracer(
// 		serviceName,
// 		jaegercfg.Logger(jLogger),
// 		jaegercfg.Metrics(jMetricsFactory),
// 		jaegercfg.Reporter(repoter),
// 	)

// 	if err != nil {
// 		log.Printf("could not initialize jaeger tracer: %s", err.Error())
// 		return nil
// 	}
// 	//defer closer.Close()
// 	return closer
// }

func Tracing(comp string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var span opentracing.Span
			opName := comp + ":" + c.Request().URL.Path
			// 监测Header中是否有Trace信息
			wireContext, err := opentracing.GlobalTracer().Extract(
				opentracing.TextMap,
				opentracing.HTTPHeadersCarrier(c.Request().Header))
			if err != nil {
				if isDebug := c.QueryParam("opentracing"); isDebug != "true" && isDebug != "1" {
					return next(c)
				}

				// 启动新Span
				span = opentracing.StartSpan(opName)
			} else {
				span = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
			}
			defer span.Finish()

			//if ctx := c.Request().Context(); ctx != nil {
			//	ctx = opentracing.ContextWithSpan(ctx, span)
			//	c.Request().WithContext(ctx)
			//}
			c.Set(DefaultKey, span)

			ext.Component.Set(span, comp)
			ext.SpanKind.Set(span, "server")
			ext.HTTPUrl.Set(span, c.Request().Host+c.Request().RequestURI)
			ext.HTTPMethod.Set(span, c.Request().Method)

			if err := next(c); err != nil {
				ext.Error.Set(span, true)
			} else {
				ext.Error.Set(span, false)
			}

			ext.HTTPStatusCode.Set(span, uint16(c.Response().Status))
			return nil
		}
	}
}

func SpanFromContext(c echo.Context) opentracing.Span {
	ot := c.Get(DefaultKey)
	if ot == nil {
		return nil
	}
	return ot.(opentracing.Span)
}
