module github.com/blcvn/backend/services/pkg

go 1.24.0

require (
	github.com/hibiken/asynq v0.24.1
	github.com/redis/go-redis/v9 v9.0.3
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.0
	google.golang.org/grpc v1.62.1
)
