package reservation

import "time"

const DefaultReservationTTL = 15 * time.Minute

type Status string

const (
	StatusPending Status = "pending"
)

type Reservation struct {
	ID         string    `json:"id"`
	ResourceID string    `json:"resource_id"`
	UserID     string    `json:"user_id"`
	Quantity   int       `json:"quantity"`
	Status     Status    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type CreateReservationRequest struct {
	ResourceID string `json:"resource_id"`
	UserID     string `json:"user_id"`
	Quantity   int    `json:"quantity"`
}
