package reservation

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const reservationIdempotencyPendingValue = "__pending__"

func normalizeIdempotencyKey(key string) string {
	return strings.TrimSpace(key)
}

func (s *ReservationService) beginCreateReservationIdempotency(ctx context.Context, key string) (*Reservation, bool, error) {
	key = normalizeIdempotencyKey(key)
	if key == "" {
		return nil, true, nil
	}

	acquired, err := s.repository.TryBeginReservationIdempotency(ctx, key, DefaultReservationTTL)
	if err != nil {
		return nil, false, err
	}

	if acquired {
		return nil, true, nil
	}

	reservationID, err := s.repository.GetReservationIDByIdempotencyKey(ctx, key)
	if err != nil {
		return nil, false, err
	}

	if reservationID == "" {
		acquired, err = s.repository.TryBeginReservationIdempotency(ctx, key, DefaultReservationTTL)
		if err != nil {
			return nil, false, err
		}

		if acquired {
			return nil, true, nil
		}

		reservationID, err = s.repository.GetReservationIDByIdempotencyKey(ctx, key)
		if err != nil {
			return nil, false, err
		}
	}

	if reservationID == reservationIdempotencyPendingValue {
		return nil, false, fmt.Errorf("%w: request already in progress for key %s", ErrReservationIdempotencyInProgress, key)
	}

	if reservationID == "" {
		return nil, false, fmt.Errorf("%w: idempotency key %s has no stored reservation", ErrReservationConflict, key)
	}

	reservation, err := s.repository.GetReservation(ctx, reservationID)
	if err != nil {
		if errors.Is(err, ErrReservationNotFound) {
			if deleteErr := s.repository.DeleteReservationIdempotency(ctx, key); deleteErr != nil {
				return nil, false, deleteErr
			}

			acquired, err = s.repository.TryBeginReservationIdempotency(ctx, key, DefaultReservationTTL)
			if err != nil {
				return nil, false, err
			}

			if acquired {
				return nil, true, nil
			}

			return nil, false, fmt.Errorf("%w: request already in progress for key %s", ErrReservationIdempotencyInProgress, key)
		}

		return nil, false, err
	}

	return reservation, false, nil
}
