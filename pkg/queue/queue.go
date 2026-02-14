package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

// RedisConfig holds connection details
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Producer wraps asynq.Client
type Producer struct {
	client *asynq.Client
}

// Consumer wraps asynq.Server
type Consumer struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewProducer initializes a new productor
func NewProducer(cfg RedisConfig) *Producer {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Producer{client: client}
}

// Enqueue sends a task to the queue
func (p *Producer) Enqueue(ctx context.Context, taskType string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(taskType, bytes)
	info, err := p.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}

func (p *Producer) Close() error {
	return p.client.Close()
}

// NewConsumer initializes a new consumer
func NewConsumer(cfg RedisConfig, concurrency int) *Consumer {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("Task %s failed: %v", task.Type(), err)
			}),
		},
	)

	return &Consumer{
		server: srv,
		mux:    asynq.NewServeMux(),
	}
}

// RegisterHandler registers a handler for a specific task type
func (c *Consumer) RegisterHandler(taskType string, handler func(context.Context, *asynq.Task) error) {
	c.mux.HandleFunc(taskType, handler)
}

// Start runs the consumer
func (c *Consumer) Start() error {
	return c.server.Run(c.mux)
}

func (c *Consumer) Stop() {
	c.server.Stop()
}
