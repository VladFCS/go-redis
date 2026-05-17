package room

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const defaultRoomCacheTTL = 5 * time.Minute

type RoomService struct {
	repository RoomRepository
	cache      RoomCache
}

func NewRoomService(repository RoomRepository, cache RoomCache) *RoomService {
	return &RoomService{
		repository: repository,
		cache:      cache,
	}
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

func (s *RoomService) GetRoomByID(ctx context.Context, req *GetRoomByIDRequest) (*Room, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is required", ErrInvalidRoom)
	}

	id := strings.TrimSpace(req.ID)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidRoom)
	}

	if s.cache != nil {
		cachedRoom, err := s.cache.GetRoom(ctx, id)
		if err == nil {
			return cachedRoom, nil
		}
	}

	resp, err := s.repository.GetRoomByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetRoom(ctx, resp, defaultRoomCacheTTL)
	}

	return resp, nil
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
