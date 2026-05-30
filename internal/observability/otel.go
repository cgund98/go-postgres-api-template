package observability

import (
	"context"
	"errors"
	golog "log"
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type OTelSetupResult struct {
	*trace.TracerProvider
	*metric.MeterProvider
	*log.LoggerProvider
}

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, serviceName string) (OTelSetupResult, func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	bootstrapLogger := NewBootstrapLogger()

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up resource metadata shared by traces, metrics, and logs.
	res, err := newResource(ctx, serviceName)
	if err != nil {
		bootstrapLogger.WarnContext(ctx, "resource detection completed with warnings", "error", err)
	}

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return OTelSetupResult{}, shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return OTelSetupResult{}, shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return OTelSetupResult{}, shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return OTelSetupResult{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
		LoggerProvider: loggerProvider,
	}, shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
}

func newTracerProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	endpoint := signalEndpoint("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(endpoint.url)}
	if endpoint.insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	traceExporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Second)),
	)
	return tracerProvider, nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	endpoint := signalEndpoint("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT")
	opts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(endpoint.hostPort)}
	if endpoint.path != "" {
		opts = append(opts, otlpmetrichttp.WithURLPath(endpoint.path))
	}
	if endpoint.insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	metricExporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context, res *resource.Resource) (*log.LoggerProvider, error) {
	endpoint := signalEndpoint("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")
	opts := []otlploghttp.Option{otlploghttp.WithEndpointURL(endpoint.url)}
	if endpoint.insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	logExporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}

type otelSignalEndpoint struct {
	url      string
	hostPort string
	path     string
	insecure bool
}

func signalEndpoint(envKey string) otelSignalEndpoint {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}
	if raw == "" {
		raw = "http://localhost:4318"
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return otelSignalEndpoint{
			url:      "http://localhost:4318",
			hostPort: "localhost:4318",
			path:     "",
			insecure: true,
		}
	}

	hostPort := parsed.Host
	if hostPort == "" {
		hostPort = "localhost:4318"
	}

	return otelSignalEndpoint{
		url:      normalizeOTelURL(parsed),
		hostPort: hostPort,
		path:     parsed.Path,
		insecure: parsed.Scheme == "http" || parsed.Scheme == "unix",
	}
}

func normalizeOTelURL(parsed *url.URL) string {
	if parsed == nil {
		return "http://localhost:4318"
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "http"
	}
	if parsed.Host == "" {
		parsed.Host = "localhost:4318"
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String()
}

var tracerCtxKey = &struct{ name string }{"otel_tracer"}

func GetTracerFromContext(ctx context.Context) oteltrace.Tracer {
	if tracer, ok := ctx.Value(tracerCtxKey).(oteltrace.Tracer); ok {
		return tracer
	}
	return nil
}

func SetTracerOnContext(ctx context.Context, tracer oteltrace.Tracer) context.Context {
	return context.WithValue(ctx, tracerCtxKey, tracer)
}

func StartTraceFromContext(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, func(options ...oteltrace.SpanEndOption)) {
	tracer := GetTracerFromContext(ctx)

	if tracer == nil {
		golog.Println("Unable to initialize trace from context. Unable to parse tracer from context")
		return ctx, func(_ ...oteltrace.SpanEndOption) {}
	}

	ctx, span := tracer.Start(ctx, spanName, opts...)
	return ctx, span.End
}
