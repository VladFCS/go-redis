package reservation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vladfc/go-redis/internal/room"
)

type ReservationService struct {
	repository ReservationRepository
	roomReader RoomReader
}

func NewReservationService(repository ReservationRepository, roomReaderRepo RoomReader) *ReservationService {
	return &ReservationService{
		repository: repository,
		roomReader: roomReaderRepo,
	}
}

func (s *ReservationService) CreateReservation(ctx context.Context, req *CreateReservationRequest, idempotencyKey string) (reservation *Reservation, err error) {
	if err := validateReservationCreateRequest(req); err != nil {
		return nil, err
	}

	idempotencyKey = normalizeIdempotencyKey(idempotencyKey)

	existingReservation, shouldCreate, err := s.beginCreateReservationIdempotency(ctx, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if !shouldCreate {
		return existingReservation, nil
	}

	if idempotencyKey != "" {
		defer func() {
			if err != nil {
				_ = s.repository.DeleteReservationIdempotency(ctx, idempotencyKey)
			}
		}()
	}

	roomID := strings.TrimSpace(req.RoomID)

	fetchedRoom, err := s.roomReader.GetRoomByID(ctx, &room.GetRoomByIDRequest{ID: roomID})
	if err != nil {
		if errors.Is(err, room.ErrRoomNotFound) {
			return nil, fmt.Errorf("%w: room_id %s not found", ErrReservationRoomNotFound, roomID)
		}

		return nil, err
	}
	if fetchedRoom.Capacity < req.Quantity {
		return nil, fmt.Errorf(
			"%w: requested quantity %d exceeds room capacity %d",
			ErrReservationCapacityExceeded,
			req.Quantity,
			fetchedRoom.Capacity,
		)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(DefaultReservationTTL)

	reservation = &Reservation{
		ID:        uuid.NewString(),
		RoomID:    roomID,
		UserID:    strings.TrimSpace(req.UserID),
		Quantity:  req.Quantity,
		Status:    ReservationStatusPending,
		CreatedAt: now,
		ExpiresAt: &expiresAt,
	}

	if err := s.repository.CreateReservation(ctx, reservation, DefaultReservationTTL, idempotencyKey); err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *ReservationService) GetReservation(ctx context.Context, reservationID string) (*Reservation, error) {
	reservationID = strings.TrimSpace(reservationID)
	if reservationID == "" {
		return nil, fmt.Errorf("%w: reservation_id is required", ErrInvalidReservation)
	}

	return s.repository.GetReservation(ctx, reservationID)
}

func (s *ReservationService) ConfirmReservation(ctx context.Context, reservationID string) (*Reservation, error) {
	reservationID = strings.TrimSpace(reservationID)
	if reservationID == "" {
		return nil, fmt.Errorf("%w: reservation_id is required", ErrInvalidReservation)
	}
	return s.repository.ConfirmReservation(ctx, reservationID)
}

func (s *ReservationService) CancelReservation(ctx context.Context, reservationID string) (*Reservation, error) {
	reservationID = strings.TrimSpace(reservationID)
	if reservationID == "" {
		return nil, fmt.Errorf("%w: reservation_id is required", ErrInvalidReservation)
	}

	return s.repository.CancelReservation(ctx, reservationID)
}

func validateReservationCreateRequest(req *CreateReservationRequest) error {
	if req == nil {
		return fmt.Errorf("%w: request is required", ErrInvalidReservation)
	}

	if strings.TrimSpace(req.RoomID) == "" {
		return fmt.Errorf("%w: room_id is required", ErrInvalidReservation)
	}

	if strings.TrimSpace(req.UserID) == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidReservation)
	}

	if req.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than 0", ErrInvalidReservation)
	}

	return nil
}
