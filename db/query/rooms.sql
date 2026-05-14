-- name: CreateRoom :exec
INSERT INTO rooms (
    id,
    room_number,
    status,
    capacity
) VALUES (
    $1,
    $2,
    $3,
    $4
);

-- name: GetRooms :many
SELECT
    id,
    room_number,
    status,
    capacity
FROM rooms
ORDER BY room_number;

-- name: GetRoomByID :one
SELECT
    id,
    room_number,
    status,
    capacity
FROM rooms
WHERE id = $1;
