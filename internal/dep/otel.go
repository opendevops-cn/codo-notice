package dep

import (
	"context"
	"net/url"
	"os"

	"codo-notice/internal/conf"
	"codo-notice/meta"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/trace"
)

func NewMeterProvider(bc *conf.Bootstrap) (metric.MeterProvider, error) {
	md := bc.Metadata
	metricConf := bc.Otel.Metric
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	if metricConf.EnableExemplar {
		err = metrics.EnableOTELExemplar()
		if err != nil {
			return nil, err
		}
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(md.Name),
				attribute.String("environment", md.Env.String()),
			),
		),
		sdkmetric.WithReader(exporter),
		sdkmetric.WithView(
			metrics.DefaultSecondsHistogramView(metrics.DefaultServerSecondsHistogramName),
		),
	)
	otel.SetMeterProvider(provider)
	return provider, nil
}

func NewTracerProvider(_ context.Context, bc *conf.Bootstrap, textMapPropagator propagation.TextMapPropagator, logger log.Logger) (trace.TracerProvider, error) {
	const (
		protocolUdp = "udp"
	)

	md := bc.Metadata
	traceConf := bc.Otel.Trace

	var exp sdktrace.SpanExporter

	u, err := url.Parse(traceConf.Endpoint)
	if err == nil && u.Scheme == protocolUdp {
		exp, err = jaeger.New(jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(u.Hostname()), jaeger.WithAgentPort(u.Port()),
		))
	} else {
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(traceConf.Endpoint)))
	}
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(md.Name),
				attribute.String("environment", md.Env.String()),
			),
		),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(textMapPropagator)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		helper := log.NewHelper(logger)
		helper.Errorf("[otel] error: %v", err)
	}))

	return tp, nil
}

func NewTextMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		tracing.Metadata{},
		propagation.Baggage{},
		propagation.TraceContext{},
	)
}

type logWrapper struct {
	logger loggersdk.Logger
}

func (x *logWrapper) Log(level log.Level, keyvals ...interface{}) error {
	switch level {
	case log.LevelDebug:
		return x.logger.Log(context.TODO(), loggersdk.LevelDebug, keyvals...)
	case log.LevelInfo:
		return x.logger.Log(context.TODO(), loggersdk.LevelInfo, keyvals...)
	case log.LevelWarn:
		return x.logger.Log(context.TODO(), loggersdk.LevelWarn, keyvals...)
	case log.LevelError:
		return x.logger.Log(context.TODO(), loggersdk.LevelError, keyvals...)
	case log.LevelFatal:
		return x.logger.Log(context.TODO(), loggersdk.LevelFatal, keyvals...)
	}

	return x.logger.Log(context.TODO(), loggersdk.LevelInfo, keyvals...)
}

func NewLogger(bc *conf.Bootstrap, logger loggersdk.Logger) (log.Logger, error) {
	md := bc.Metadata
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return log.With(
		&logWrapper{logger: logger},
		"caller", log.DefaultCaller,
		"service.id", hostname,
		"service.name", md.Name,
		"service.version", meta.Version,
		"trace", tracing.TraceID(),
		"span", tracing.SpanID(),
	), nil
}
