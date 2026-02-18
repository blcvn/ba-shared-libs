package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/blcvn/backend/services/ba-agent-service/domain"
	"github.com/go-redis/redis/v8"
)

// CachedLLM wraps an LLM service with Redis caching
type CachedLLM struct {
	delegate        domain.LLMService
	redisClient     *redis.Client
	ttl             time.Duration
	modelIdentifier string // Cache isolation by model
}

// NewCachedLLM creates a new cached LLM service
func NewCachedLLM(delegate domain.LLMService, redisClient *redis.Client, ttl time.Duration, modelIdentifier string) *CachedLLM {
	return &CachedLLM{
		delegate:        delegate,
		redisClient:     redisClient,
		ttl:             ttl,
		modelIdentifier: modelIdentifier,
	}
}

// Chat implements the domain.LLMService interface with caching
func (c *CachedLLM) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// 1. Generate Cache Key
	hash := sha256.New()
	hash.Write([]byte(c.modelIdentifier))
	hash.Write([]byte(systemPrompt))
	hash.Write([]byte(userPrompt))
	key := fmt.Sprintf("llm:cache:%s", hex.EncodeToString(hash.Sum(nil)))

	// 2. Check Cache
	val, err := c.redisClient.Get(ctx, key).Result()
	if err == nil {
		log.Printf("DEBUG: LLM Cache HIT for key %s", key)
		return val, nil
	} else if err != redis.Nil {
		log.Printf("WARN: Redis error: %v", err)
	} else {
		log.Printf("DEBUG: LLM Cache MISS for key %s", key)
	}

	// 3. Call Delegate
	resp, err := c.delegate.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	// 4. Cache Result
	if err := c.redisClient.Set(ctx, key, resp, c.ttl).Err(); err != nil {
		log.Printf("WARN: Failed to set cache: %v", err)
	}

	return resp, nil
}
