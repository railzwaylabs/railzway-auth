package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/smallbiznis/railzway-auth/internal/domain/oauth"
	"github.com/smallbiznis/railzway-auth/internal/repository"
)

// RedisAuthorizeStateStore implements AuthorizeStateStore backed by Redis.
type RedisAuthorizeStateStore struct {
	client redis.UniversalClient
}

var _ repository.AuthorizeStateStore = (*RedisAuthorizeStateStore)(nil)

// NewRedisAuthorizeStateStore constructs a Redis-backed authorize state store.
func NewRedisAuthorizeStateStore(client redis.UniversalClient) *RedisAuthorizeStateStore {
	return &RedisAuthorizeStateStore{client: client}
}

// SaveState stores the encoded authorize state payload with TTL.
func (s *RedisAuthorizeStateStore) SaveState(ctx context.Context, key string, data oauth.AuthorizeState, ttl time.Duration) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal authorize state: %w", err)
	}
	if err := s.client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("persist authorize state: %w", err)
	}
	return nil
}

// GetState loads and decodes the authorize state payload.
func (s *RedisAuthorizeStateStore) GetState(ctx context.Context, key string) (*oauth.AuthorizeState, error) {
	bytes, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("load authorize state: %w", err)
	}
	var state oauth.AuthorizeState
	if err := json.Unmarshal(bytes, &state); err != nil {
		return nil, fmt.Errorf("decode authorize state: %w", err)
	}
	return &state, nil
}

// DeleteState removes the persisted authorize state key.
func (s *RedisAuthorizeStateStore) DeleteState(ctx context.Context, key string) error {
	if err := s.client.Del(ctx, key).Err(); err != nil && err != redis.Nil {
		return fmt.Errorf("delete authorize state: %w", err)
	}
	return nil
}
