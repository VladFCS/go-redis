CREATE TABLE rooms (
    id TEXT PRIMARY KEY,
    room_number TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'available',
    capacity INTEGER NOT NULL CHECK (capacity > 0)
);
