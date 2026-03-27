package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config holds telemetry configuration.
type Config struct {
	ServiceName  string
	CollectorURL string
	Enabled      bool
	Insecure     bool
}

// Provider encapsulates tracing and metrics providers.
type Provider struct {
	tp          *sdktrace.TracerProvider
	mp          *sdkmetric.MeterProvider
	httpMetrics *HTTPMetrics
}

// Setup initializes OpenTelemetry (Traces + Metrics).
// Returns a Provider that must be shut down on application exit.
func Setup(ctx context.Context, cfg Config) (*Provider, error) {
	if !cfg.Enabled || cfg.CollectorURL == "" {
		slog.Info("OpenTelemetry disabled or no collector URL configured")
		return &Provider{}, nil
	}

	// Resource with service information
	res, resErr := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if resErr != nil {
		return nil, resErr
	}

	// Traces
	tp, tpErr := setupTracer(ctx, cfg.CollectorURL, cfg.Insecure, res)
	if tpErr != nil {
		return nil, tpErr
	}

	// Metrics
	mp, mpErr := setupMeter(ctx, cfg.CollectorURL, cfg.Insecure, res)
	if mpErr != nil {
		return nil, mpErr
	}

	// HTTP Metrics
	httpMetrics, httpErr := NewHTTPMetrics(cfg.ServiceName)
	if httpErr != nil {
		return nil, httpErr
	}

	slog.Info("OpenTelemetry initialized",
		"service", cfg.ServiceName,
		"collector", cfg.CollectorURL,
	)

	return &Provider{tp: tp, mp: mp, httpMetrics: httpMetrics}, nil
}

func setupTracer(ctx context.Context, collectorURL string, insecure bool, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	traceOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(collectorURL),
	}
	if insecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
	}

	exporter, exporterErr := otlptracegrpc.New(ctx, traceOpts...)
	if exporterErr != nil {
		return nil, exporterErr
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

func setupMeter(ctx context.Context, collectorURL string, insecure bool, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	metricOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(collectorURL),
	}
	if insecure {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}

	exporter, exporterErr := otlpmetricgrpc.New(ctx, metricOpts...)
	if exporterErr != nil {
		return nil, exporterErr
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(mp)

	return mp, nil
}

// Shutdown shuts down telemetry providers gracefully.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tp != nil {
		if tpErr := p.tp.Shutdown(ctx); tpErr != nil {
			return tpErr
		}
	}
	if p.mp != nil {
		if mpErr := p.mp.Shutdown(ctx); mpErr != nil {
			return mpErr
		}
	}
	return nil
}

// HTTPMetrics returns the HTTP metrics instance.
func (p *Provider) HTTPMetrics() *HTTPMetrics {
	return p.httpMetrics
}
