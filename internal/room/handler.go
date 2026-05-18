package room

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *RoomService
}

func NewHandler(service *RoomService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	router := chi.NewRouter()

	router.Route("/rooms", func(r chi.Router) {
		r.Get("/", h.GetRooms)
		r.Get("/{id}", h.GetRoomByID)
		r.Post("/", h.CreateRoom)
	})

	return router
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var request CreateRoomRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	room, err := h.service.CreateRoom(ctx, &request)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, room)
}

func (h *Handler) GetRoomByID(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	room, err := h.service.GetRoomByID(ctx, &GetRoomByIDRequest{ID: id})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, room)
}

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rooms, err := h.service.GetRooms(ctx, &GetRoomsRequest{})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, rooms)
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrInvalidRoom) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrRoomNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrRoomConflict) {
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
