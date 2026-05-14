package room

import (
	"context"
	"fmt"
	"strings"
)

type RoomService struct {
	repository RoomRepository
}

func NewRoomService(repository RoomRepository) *RoomService {
	return &RoomService{repository: repository}
}

func (s *RoomService) CreateRoom(ctx context.Context, req *CreateRoomRequest) (*Room, error) {
	if err := validateRoomCreateRequest(req); err != nil {
		return nil, err
	}

	room := &Room{
		ID:         strings.TrimSpace(req.ID),
		RoomNumber: strings.TrimSpace(req.RoomNumber),
		Status:     RoomStatusAvailable,
		Capacity:   req.Capacity,
	}

	if err := s.repository.CreateRoom(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func validateRoomCreateRequest(req *CreateRoomRequest) error {
	if req == nil {
		return fmt.Errorf("%w: request is required", ErrInvalidRoom)
	}

	if strings.TrimSpace(req.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidRoom)
	}

	if strings.TrimSpace(req.RoomNumber) == "" {
		return fmt.Errorf("%w: room_number is required", ErrInvalidRoom)
	}

	if req.Capacity <= 0 || req.Capacity > 4 {
		return fmt.Errorf("%w: capacity must be between 1 and 4", ErrInvalidRoom)
	}

	return nil
}
