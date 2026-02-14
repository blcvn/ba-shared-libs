package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns an interceptor that extracts trace context from metadata
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// Extract context from metadata
			propagator := otel.GetTextMapPropagator()
			// metadata.MD implements TextMapCarrier indirectly via simple wrapper if needed
			// For simplicity, we just log trace ID if present
			log.Printf("[Trace] Metadata: %v", md)

			// To properly propagate:
			carrier := metadataCarrier(md)
			ctx = propagator.Extract(ctx, carrier)
		}

		tracer := otel.Tracer("grpc-interceptor")
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		log.Printf("[Trace] Method: %s TraceID: %s", info.FullMethod, span.SpanContext().TraceID())

		return handler(ctx, req)
	}
}

type metadataCarrier metadata.MD

func (m metadataCarrier) Get(key string) string {
	vals := metadata.MD(m).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (m metadataCarrier) Set(key string, value string) {
	metadata.MD(m).Set(key, value)
}

func (m metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
