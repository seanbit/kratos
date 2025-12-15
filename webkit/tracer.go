package webkit

import (
	"context"
	"fmt"

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
	if host == "" && port == 0 {
		log.Infow("msg", "trace disabled")
		return nil
	}

	var err error
	var exp tracesdk.SpanExporter
	switch providerType {
	case "jaeger":
		// Create the Jaeger exporter
		exp, err = jaeger.New(jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(host),
			jaeger.WithAgentPort(fmt.Sprint(port)),
		))
	case "OTLP":
		clientOpts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", host, port)),
			otlptracegrpc.WithInsecure(),
		}
		exp, err = otlptrace.New(context.Background(), otlptracegrpc.NewClient(clientOpts...))
	}
	if err != nil {
		return err
	}

	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(1.0))),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
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
