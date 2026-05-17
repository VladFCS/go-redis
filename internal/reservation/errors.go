package reservation

import "errors"

var (
	ErrInvalidReservation          = errors.New("invalid reservation")
	ErrReservationNotFound         = errors.New("reservation not found")
	ErrReservationConflict         = errors.New("reservation conflict")
	ErrReservationRoomNotFound     = errors.New("reservation room not found")
	ErrReservationCapacityExceeded = errors.New("reservation capacity exceeded")
)
