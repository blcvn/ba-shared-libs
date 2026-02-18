package queue

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const TaskQueueKey = "ba_agent:task_queue"

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

// Publish adds a task ID to the queue
func (q *RedisQueue) Publish(ctx context.Context, taskID string) error {
	return q.client.RPush(ctx, TaskQueueKey, taskID).Err()
}

// Consume implements a blocking pop to get tasks
func (q *RedisQueue) Consume(ctx context.Context) (string, error) {
	// BLPop returns [key, value]
	result, err := q.client.BLPop(ctx, 0*time.Second, TaskQueueKey).Result()
	if err != nil {
		return "", err
	}
	if len(result) < 2 {
		return "", nil // Should not happen with BLPop
	}
	return result[1], nil
}
