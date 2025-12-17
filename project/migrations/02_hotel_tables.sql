CREATE TABLE hotels (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    contact_phone TEXT NOT NULL
);

CREATE TABLE room_types_in_hotels (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER REFERENCES hotels(id),
    type TEXT NOT NULL,
    price_per_night DECIMAL(10,2) NOT NULL,
    max_guests INTEGER DEFAULT 2
);

CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    room_type INTEGER REFERENCES room_types_in_hotels(id),
    room_number TEXT NOT NULL
);

CREATE INDEX idx_rooms_type_hotel ON room_types_in_hotels(hotel_id);
CREATE INDEX idx_rooms_hotel ON rooms(room_type);
