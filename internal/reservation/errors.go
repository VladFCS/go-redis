package reservation

import "errors"

var (
	ErrInvalidReservation  = errors.New("invalid reservation")
	ErrReservationNotFound = errors.New("reservation not found")
)
