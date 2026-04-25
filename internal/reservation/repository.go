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
	ConfirmReservation(ctx context.Context, id string) (*Reservation, error)
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
	})
	if reservation.ExpiresAt != nil {
		pipe.HSet(ctx, key, "expires_at", reservation.ExpiresAt.UTC().Format(time.RFC3339))
	}

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

	return reservationFromHash(result)
}

func (r *RedisRepository) ConfirmReservation(ctx context.Context, id string) (*Reservation, error) {
	key := reservationKey(id)

	for range 3 {
		var confirmedReservation *Reservation

		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			result, err := tx.HGetAll(ctx, key).Result()
			if err != nil {
				return err
			}

			if len(result) == 0 {
				return ErrReservationNotFound
			}

			reservation, err := reservationFromHash(result)
			if err != nil {
				return err
			}

			if reservation.Status != StatusPending {
				return fmt.Errorf("cannot confirm reservation with status %s", reservation.Status)
			}

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.HSet(ctx, key, "status", StatusConfirmed)
				pipe.HDel(ctx, key, "expires_at")
				pipe.Persist(ctx, key)
				return nil
			})
			if err != nil {
				return err
			}

			reservation.Status = StatusConfirmed
			reservation.ExpiresAt = nil
			confirmedReservation = reservation

			return nil
		}, key)
		if err == redis.TxFailedErr {
			continue
		}

		return confirmedReservation, err
	}

	return nil, redis.TxFailedErr
}

func reservationFromHash(result map[string]string) (*Reservation, error) {
	quantity, err := strconv.Atoi(result["quantity"])
	if err != nil {
		return nil, fmt.Errorf("parse quantity: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, result["created_at"])
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	var expiresAt *time.Time
	if value := result["expires_at"]; value != "" {
		parsedExpiresAt, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, fmt.Errorf("parse expires_at: %w", err)
		}
		expiresAt = &parsedExpiresAt
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
