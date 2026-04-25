package reservation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateReservation(ctx context.Context, req CreateReservationRequest) (*Reservation, error) {
	if err := validateCreateReservationRequest(req); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(DefaultReservationTTL)

	reservation := &Reservation{
		ID:         uuid.NewString(),
		ResourceID: strings.TrimSpace(req.ResourceID),
		UserID:     strings.TrimSpace(req.UserID),
		Quantity:   req.Quantity,
		Status:     StatusPending,
		CreatedAt:  now,
		ExpiresAt:  &expiresAt,
	}

	if err := s.repository.CreateReservation(ctx, reservation, DefaultReservationTTL); err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *Service) ConfirmReservation(ctx context.Context, reservationID string) (*Reservation, error) {
	return s.repository.ConfirmReservation(ctx, reservationID)
}

func validateCreateReservationRequest(req CreateReservationRequest) error {
	if strings.TrimSpace(req.ResourceID) == "" {
		return fmt.Errorf("%w: resource_id is required", ErrInvalidReservation)
	}

	if strings.TrimSpace(req.UserID) == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidReservation)
	}

	if req.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than 0", ErrInvalidReservation)
	}

	return nil
}
