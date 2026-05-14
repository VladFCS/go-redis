package room

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	roompostgres "github.com/vladfc/go-redis/internal/room/postgres"
)

type RoomRepository interface {
	CreateRoom(ctx context.Context, room *Room) error
	GetRooms(ctx context.Context) ([]*Room, error)
	GetRoomByID(ctx context.Context, id string) (*Room, error)
}

type PostgreSQLRepository struct {
	client *sql.DB
	query  *roompostgres.Queries
}

func NewPostgreSQLRepository(client *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{
		client: client,
		query:  roompostgres.New(client),
	}
}

func (r *PostgreSQLRepository) CreateRoom(ctx context.Context, room *Room) error {
	err := r.query.CreateRoom(ctx, roompostgres.CreateRoomParams{
		ID:         room.ID,
		RoomNumber: room.RoomNumber,
		Status:     string(room.Status),
		Capacity:   int32(room.Capacity),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrRoomConflict
		}

		return err
	}

	return nil
}

func (r *PostgreSQLRepository) GetRooms(ctx context.Context) ([]*Room, error) {
	rows, err := r.query.GetRooms(ctx)
	if err != nil {
		return nil, err
	}

	rooms := make([]*Room, 0, len(rows))
	for _, row := range rows {
		rooms = append(rooms, &Room{
			ID:         row.ID,
			RoomNumber: row.RoomNumber,
			Status:     RoomStatus(row.Status),
			Capacity:   int(row.Capacity),
		})
	}

	return rooms, nil
}

func (r *PostgreSQLRepository) GetRoomByID(ctx context.Context, id string) (*Room, error) {
	row, err := r.query.GetRoomByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoomNotFound
		}

		return nil, err
	}

	return &Room{
		ID:         row.ID,
		RoomNumber: row.RoomNumber,
		Status:     RoomStatus(row.Status),
		Capacity:   int(row.Capacity),
	}, nil
}
