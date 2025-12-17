package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Booking struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	HotelID      int       `json:"hotel_id"`
	RoomID       int       `json:"room_id"`
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
	GuestsCount  int       `json:"guests_count"`
	TotalPrice   float64   `json:"total_price"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateBooking(ctx context.Context, booking *Booking) (int, error) {
	query := `
		INSERT INTO bookings 
		(user_id, hotel_id, room_id, check_in_date, check_out_date, guests_count, total_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		booking.UserID,
		booking.HotelID,
		booking.RoomID,
		booking.CheckInDate,
		booking.CheckOutDate,
		booking.GuestsCount,
		booking.TotalPrice,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create booking: %w", err)
	}

	return id, nil
}

func (r *Repository) GetUserBookings(ctx context.Context, userID int) ([]Booking, error) {
	query := `
		SELECT id, user_id, hotel_id, room_id, check_in_date, check_out_date, 
		       guests_count, total_price
		FROM bookings 
		WHERE user_id = $1
		ORDER BY check_in_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user bookings: %w", err)
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		err := rows.Scan(
			&b.ID,
			&b.UserID,
			&b.HotelID,
			&b.RoomID,
			&b.CheckInDate,
			&b.CheckOutDate,
			&b.GuestsCount,
			&b.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

func (r *Repository) CheckRoomAvailability(ctx context.Context, roomID int, checkIn, checkOut time.Time) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM bookings 
			WHERE room_id = $1 
			AND NOT (check_out_date <= $2 OR check_in_date >= $3)
		)
	`

	var isOccupied bool
	err := r.db.QueryRowContext(ctx, query, roomID, checkIn, checkOut).Scan(&isOccupied)
	if err != nil {
		return false, fmt.Errorf("failed to check availability: %w", err)
	}

	return !isOccupied, nil
}

func (r *Repository) GetBusyRooms(ctx context.Context, checkIn, checkOut time.Time) (map[int]bool, error) {
	query := `
        SELECT room_id 
        FROM bookings 
        WHERE NOT (check_out_date <= $1 OR check_in_date >= $2)
    `
	rows, err := r.db.QueryContext(ctx, query, checkIn, checkOut)
	if err != nil {
		return nil, fmt.Errorf("failed to get busy rooms: %w", err)
	}
	defer rows.Close()

	busyRooms := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		busyRooms[id] = true
	}
	return busyRooms, nil
}

func (r *Repository) GetHotelBookings(ctx context.Context, hotelID int) ([]Booking, error) {
	query := `
		SELECT id, user_id, hotel_id, room_id, check_in_date, check_out_date, 
		       guests_count, total_price
		FROM bookings 
		WHERE hotel_id = $1
		ORDER BY check_in_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, hotelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query hotel bookings: %w", err)
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		err := rows.Scan(
			&b.ID,
			&b.UserID,
			&b.HotelID,
			&b.RoomID,
			&b.CheckInDate,
			&b.CheckOutDate,
			&b.GuestsCount,
			&b.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}
