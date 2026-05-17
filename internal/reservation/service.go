package reservation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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

func (s *ReservationService) CreateReservation(ctx context.Context, req *CreateReservationRequest) (*Reservation, error) {
	if err := validateReservationCreateRequest(req); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(DefaultReservationTTL)

	reservation := &Reservation{
		ID:        uuid.NewString(),
		RoomID:    strings.TrimSpace(req.RoomID),
		UserID:    strings.TrimSpace(req.UserID),
		Quantity:  req.Quantity,
		Status:    ReservationStatusPending,
		CreatedAt: now,
		ExpiresAt: &expiresAt,
	}

	if err := s.repository.CreateReservation(ctx, reservation, DefaultReservationTTL); err != nil {
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
