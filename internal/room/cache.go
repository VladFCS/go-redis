package room

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var errRoomCacheMiss = errors.New("room cache miss")

type RoomCache interface {
	GetRoom(ctx context.Context, id string) (*Room, error)
	SetRoom(ctx context.Context, room *Room, ttl time.Duration) error
}

type RedisRoomCache struct {
	client *redis.Client
}

func NewRedisRoomCache(client *redis.Client) *RedisRoomCache {
	return &RedisRoomCache{client: client}
}

func (c *RedisRoomCache) GetRoom(ctx context.Context, id string) (*Room, error) {
	payload, err := c.client.Get(ctx, roomKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errRoomCacheMiss
		}

		return nil, err
	}

	var room Room
	if err := json.Unmarshal(payload, &room); err != nil {
		return nil, err
	}

	return &room, nil
}

func (c *RedisRoomCache) SetRoom(ctx context.Context, room *Room, ttl time.Duration) error {
	payload, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, roomKey(room.ID), payload, ttl).Err()
}
