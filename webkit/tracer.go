package webkit

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// InitTracerProvider  Set global trace provider
func InitTracerProvider(providerType, host string, port int, serviceName, serviceVersion string, env string) error {
	return InitTracerProviderWithSampling(providerType, host, port, serviceName, serviceVersion, env, 1.0)
}

// InitTracerProviderWithSampling sets global trace provider with configurable sampling rate.
// samplingRate: 0.0 to 1.0, where 1.0 means 100% sampling.
func InitTracerProviderWithSampling(providerType, host string, port int, serviceName, serviceVersion string, env string, samplingRate float64) error {
	if host == "" && port == 0 {
		log.Infow("msg", "trace disabled")
		return nil
	}

	var err error
	var exp tracesdk.SpanExporter
	switch strings.ToLower(providerType) {
	case "jaeger":
		// Create the Jaeger exporter
		exp, err = jaeger.New(jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(host),
			jaeger.WithAgentPort(fmt.Sprint(port)),
		))
	case "otlp":
		clientOpts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", host, port)),
			otlptracegrpc.WithInsecure(),
		}
		exp, err = otlptrace.New(context.Background(), otlptracegrpc.NewClient(clientOpts...))
	default:
		return fmt.Errorf("unsupported tracer provider type: %q, expected \"jaeger\" or \"otlp\"", providerType)
	}
	if err != nil {
		return err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(samplingRate))),
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("env", env),
		)),
	)
	otel.SetTracerProvider(tp)

	log.Infof("InitTracerProvider, type: %s, host: %s, port: %d", providerType, host, port)
	return nil
}
