package room

import "errors"

var (
	ErrInvalidRoom  = errors.New("invalid room")
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomConflict = errors.New("room conflict")
)
