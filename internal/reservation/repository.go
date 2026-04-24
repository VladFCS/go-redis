package reservation

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	CreateReservation(ctx context.Context, reservation *Reservation, ttl time.Duration) error
	GetReservation(ctx context.Context, id string) (*Reservation, error)
	ConfirmReservation(ctx context.Context, id string) error
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

func (r *RedisRepository) GetReservation(ctx context.Context, id string) (*Reservation, error) {
	key := reservationKey(id)
	result, err := r.client.HGetAll(ctx, key).Result()

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrReservationNotFound
	}

	quantity, err := strconv.Atoi(result["quantity"])
	if err != nil {
		return nil, fmt.Errorf("parse quantity: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, result["created_at"])
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	expiresAt, err := time.Parse(time.RFC3339, result["expires_at"])
	if err != nil {
		return nil, fmt.Errorf("parse expires_at: %w", err)
	}

	reservation := &Reservation{
		ID:         result["id"],
		ResourceID: result["resource_id"],
		UserID:     result["user_id"],
		Quantity:   quantity,
		Status:     Status(result["status"]),
		CreatedAt:  createdAt,
		ExpiresAt:  expiresAt,
	}

	return reservation, nil
}

func (r *RedisRepository) ConfirmReservation(ctx context.Context, id string) error {
	key := reservationKey(id)

	for range 3 {
		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			result, err := tx.HGetAll(ctx, key).Result()
			if err != nil {
				return err
			}

			if len(result) == 0 {
				return ErrReservationNotFound
			}

			status := Status(result["status"])
			if status != StatusPending {
				return fmt.Errorf("cannot confirm reservation with status %s", status)
			}

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.HSet(ctx, key, map[string]any{
					"status": StatusConfirmed,
				})
				pipe.Persist(ctx, key)
				return nil
			})

			return err
		}, key)
		if err == redis.TxFailedErr {
			continue
		}

		return err
	}

	return redis.TxFailedErr
}
