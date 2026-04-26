package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	router := chi.NewRouter()

	router.Route("/reservations", func(r chi.Router) {
		r.Get("/{id}", h.GetReservation)
		r.Post("/", h.CreateReservation)
		r.Post("/{id}/confirm", h.ConfirmReservation)
	})

	return router
}

func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var request CreateReservationRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	reservation, err := h.service.CreateReservation(ctx, request)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, reservation)
}

func (h *Handler) ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	reservationID := chi.URLParam(r, "id")
	if reservationID == "" {
		writeError(w, http.StatusBadRequest, "reservation_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	confirmReservation, err := h.service.ConfirmReservation(ctx, reservationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, confirmReservation)
}

func (h *Handler) GetReservation(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	reservationID := chi.URLParam(r, "id")
	if reservationID == "" {
		writeError(w, http.StatusBadRequest, "reservation_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	reservation, err := h.service.GetReservation(ctx, reservationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, reservation)
}
 
func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrInvalidReservation) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrReservationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrReservationConflict) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeError(w, http.StatusInternalServerError, "internal server error")
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}
