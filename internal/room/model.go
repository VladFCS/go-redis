package room

type RoomStatus string

const (
	RoomStatusAvailable RoomStatus = "available"
	RoomStatusReserved  RoomStatus = "reserved"
)

type Room struct {
	ID         string     `json:"id"`
	RoomNumber string     `json:"room_number"`
	Status     RoomStatus `json:"status"`
	Capacity   int        `json:"capacity"`
}

type CreateRoomRequest struct {
	ID         string `json:"id"`
	RoomNumber string `json:"room_number"`
	Capacity   int    `json:"capacity"`
}

type GetRoomsRequest struct {
	Status RoomStatus `json:"status"`
}

type GetRoomByIDRequest struct {
	ID string `json:"id"`
}
