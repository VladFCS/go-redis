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
	ResourceID string            `json:"resource_id"`
	UserID     string            `json:"user_id"`
	Quantity   int               `json:"quantity"`
	Status     ReservationStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"`
}

type CreateReservationRequest struct {
	ResourceID string `json:"resource_id"`
	UserID     string `json:"user_id"`
	Quantity   int    `json:"quantity"`
}

type ConfirmReservationRequest struct {
	ResourceID string `json:"resource_id"`
	UserID     string `json:"user_id"`
}
