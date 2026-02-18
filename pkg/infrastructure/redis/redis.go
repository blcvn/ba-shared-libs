package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/blcvn/backend/services/pkg/entities"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (r *RedisClient) SaveSession(ctx context.Context, session *entities.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "session:"+session.ID, data, 24*time.Hour).Err()
}

func (r *RedisClient) GetSession(ctx context.Context, id string) (*entities.Session, error) {
	data, err := r.client.Get(ctx, "session:"+id).Bytes()
	if err != nil {
		return nil, err
	}
	var session entities.Session
	err = json.Unmarshal(data, &session)
	return &session, err
}

func (r *RedisClient) AddStep(ctx context.Context, sessionID string, step entities.Step) error {
	session, err := r.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	session.History = append(session.History, step)
	session.UpdatedAt = time.Now()
	return r.SaveSession(ctx, session)
}
