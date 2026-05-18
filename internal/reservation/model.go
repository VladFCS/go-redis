package reservation

import "time"

const DefaultReservationTTL = 15 * time.Minute

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
)

type Reservation struct {
	ID         string            `json:"id"`
	RoomID     string            `json:"room_id"`
	UserID     string            `json:"user_id"`
	Quantity   int               `json:"quantity"`
	Status     ReservationStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"`
	TTLSeconds *int64            `json:"ttl_seconds,omitempty"`
}

type CreateReservationRequest struct {
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Quantity int    `json:"quantity"`
}

type ConfirmReservationRequest struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}
