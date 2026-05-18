package reservation

func reservationKey(id string) string {
	return "reservation:" + id
}

func reservationIdempotencyKey(key string) string {
	return "idempotency:reservation:" + key
}
