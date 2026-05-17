package reservation

import (
	"context"

	"github.com/vladfc/go-redis/internal/room"
)

type RoomReader interface {
	GetRoomByID(ctx context.Context, req *room.GetRoomByIDRequest) (*room.Room, error)
}
