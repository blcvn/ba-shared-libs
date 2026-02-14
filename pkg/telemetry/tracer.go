package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer(serviceName string) *trace.TracerProvider {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// No-op exporter for now. In prod, use OTLP/Jaeger.
	tp := trace.NewTracerProvider(
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp
}
