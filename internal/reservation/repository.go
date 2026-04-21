package reservation

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	CreateReservation(ctx context.Context, reservation *Reservation, ttl time.Duration) error
}

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) CreateReservation(ctx context.Context, reservation *Reservation, ttl time.Duration) error {
	key := reservationKey(reservation.ID)

	pipe := r.client.TxPipeline()

	pipe.HSet(ctx, key, map[string]any{
		"id":          reservation.ID,
		"resource_id": reservation.ResourceID,
		"user_id":     reservation.UserID,
		"quantity":    reservation.Quantity,
		"status":      reservation.Status,
		"created_at":  reservation.CreatedAt.UTC().Format(time.RFC3339),
		"expires_at":  reservation.ExpiresAt.UTC().Format(time.RFC3339),
	})

	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	return err
}
