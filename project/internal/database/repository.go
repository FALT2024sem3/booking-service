package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound = errors.New("record not found")
)

// === Models for database ===

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

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

type Hotel struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	ContactPhone string `json:"contact_phone"`
}

type RoomType struct {
	ID            int     `json:"id"`
	HotelID       int     `json:"hotel_id"`
	Type          string  `json:"type"`
	PricePerNight float64 `json:"price_per_night"`
	MaxGuests     int     `json:"max_guests"`
}

type Room struct {
	ID          int    `json:"id"`
	RoomType    int    `json:"room_type"`
	RoomNumber  string `json:"room_number"`
	IsAvailable bool   `json:"is_available"`
}

// === Booking Repository ===

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create new booking
func (r *BookingRepository) CreateBooking(ctx context.Context, booking *Booking) (int, error) {
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

	booking.ID = id
	return id, nil
}

// GetBookingByID return booking by ID
func (r *BookingRepository) GetBookingByID(ctx context.Context, id int) (*Booking, error) {
	query := `
		SELECT id, user_id, hotel_id, room_id, check_in_date, check_out_date, 
		       guests_count, total_price
		FROM bookings 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var booking Booking
	err := row.Scan(
		&booking.ID,
		&booking.UserID,
		&booking.HotelID,
		&booking.RoomID,
		&booking.CheckInDate,
		&booking.CheckOutDate,
		&booking.GuestsCount,
		&booking.TotalPrice,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	return &booking, nil
}

// GetUserBookings return all user's bookings
func (r *BookingRepository) GetUserBookings(ctx context.Context, userID int) ([]Booking, error) {
	query := `
		SELECT id, user_id, hotel_id, room_id, check_in_date, check_out_date, 
		       guests_count, total_price
		FROM bookings 
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookings: %w", err)
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var booking Booking
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.HotelID,
			&booking.RoomID,
			&booking.CheckInDate,
			&booking.CheckOutDate,
			&booking.GuestsCount,
			&booking.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, booking)
	}

	return bookings, nil
}

// CheckRoomAvailability checks room availability
func (r *BookingRepository) CheckRoomAvailability(ctx context.Context, hotelID, roomID int, checkIn, checkOut time.Time) (bool, error) {
	query := `
		SELECT NOT EXISTS (
			SELECT 1 FROM bookings 
			WHERE hotel_id = $1 AND room_id = $2
			AND check_in_date <= $4 AND check_out_date >= $3
		) as is_available
	`

	var isAvailable bool
	err := r.db.QueryRowContext(ctx, query, hotelID, roomID, checkIn, checkOut).Scan(&isAvailable)
	if err != nil {
		return false, fmt.Errorf("failed to check availability: %w", err)
	}

	return isAvailable, nil
}

// === User Repository ===

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates new user
func (r *UserRepository) CreateUser(ctx context.Context, user *User) (int, error) {
	query := `
		INSERT INTO users (email, full_name, phone)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.FullName,
		user.Phone,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = id
	return id, nil
}

// GetUserByID return user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, email, full_name, phone
		FROM users 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.Phone,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail return user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, full_name, phone
		FROM users 
		WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.Phone,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// === Hotel Repository ===

type HotelRepository struct {
	db *sql.DB
}

func NewHotelRepository(db *sql.DB) *HotelRepository {
	return &HotelRepository{db: db}
}

// GetAllHotels returns all hotels
func (r *HotelRepository) GetAllHotels(ctx context.Context) ([]Hotel, error) {
	query := `
		SELECT id, name, address, contact_phone
		FROM hotels
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query hotels: %w", err)
	}
	defer rows.Close()

	var hotels []Hotel
	for rows.Next() {
		var hotel Hotel
		err := rows.Scan(
			&hotel.ID,
			&hotel.Name,
			&hotel.Address,
			&hotel.ContactPhone,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hotel: %w", err)
		}
		hotels = append(hotels, hotel)
	}

	return hotels, nil
}

// GetHotelByID returns hotel by ID
func (r *HotelRepository) GetHotelByID(ctx context.Context, id int) (*Hotel, error) {
	query := `
		SELECT id, name, address, contact_phone
		FROM hotels 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var hotel Hotel
	err := row.Scan(
		&hotel.ID,
		&hotel.Name,
		&hotel.Address,
		&hotel.ContactPhone,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get hotel: %w", err)
	}

	return &hotel, nil
}

// GetRoomTypes returns all room types
func (r *HotelRepository) GetRoomTypes(ctx context.Context, hotelID int) ([]RoomType, error) {
	query := `
		SELECT id, hotel_id, type, price_per_night, max_guests
		FROM room_types_in_hotels 
		WHERE hotel_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, hotelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query room types: %w", err)
	}
	defer rows.Close()

	var roomTypes []RoomType
	for rows.Next() {
		var rt RoomType
		err := rows.Scan(
			&rt.ID,
			&rt.HotelID,
			&rt.Type,
			&rt.PricePerNight,
			&rt.MaxGuests,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room type: %w", err)
		}
		roomTypes = append(roomTypes, rt)
	}

	return roomTypes, nil
}

// GetRoomTypeByID returns room type by ID
func (r *HotelRepository) GetRoomTypeByID(ctx context.Context, id int) (*RoomType, error) {
	query := `
		SELECT id, hotel_id, type, price_per_night, max_guests
		FROM room_types_in_hotels 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var roomType RoomType
	err := row.Scan(
		&roomType.ID,
		&roomType.HotelID,
		&roomType.Type,
		&roomType.PricePerNight,
		&roomType.MaxGuests,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room type: %w", err)
	}

	return &roomType, nil
}

// GetRoomsByType returns all rooms by type
func (r *HotelRepository) GetRoomsByType(ctx context.Context, roomType int) ([]Room, error) {
	query := `
		SELECT id, room_type, room_number, is_available
		FROM rooms 
		WHERE room_type = $1
		ORDER BY room_number
	`

	rows, err := r.db.QueryContext(ctx, query, roomType)
	if err != nil {
		return nil, fmt.Errorf("failed to query rooms: %w", err)
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(
			&room.ID,
			&room.RoomType,
			&room.RoomNumber,
			&room.IsAvailable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// GetAvailableRooms returns all available rooms
func (r *HotelRepository) GetAvailableRooms(ctx context.Context, roomType int) ([]Room, error) {
	query := `
		SELECT id, room_type, room_number, is_available
		FROM rooms 
		WHERE room_type = $1 AND is_available = true
		ORDER BY room_number
	`

	rows, err := r.db.QueryContext(ctx, query, roomType)
	if err != nil {
		return nil, fmt.Errorf("failed to query rooms: %w", err)
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(
			&room.ID,
			&room.RoomType,
			&room.RoomNumber,
			&room.IsAvailable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// GetHotelBookings return all hotel's bookings
func (r *BookingRepository) GetHotelBookings(ctx context.Context, hotelID int) ([]Booking, error) {
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
		var booking Booking
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.HotelID,
			&booking.RoomID,
			&booking.CheckInDate,
			&booking.CheckOutDate,
			&booking.GuestsCount,
			&booking.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, booking)
	}

	return bookings, nil
}

// UpdateRoomAvailability update room availability
func (r *HotelRepository) UpdateRoomAvailability(ctx context.Context, roomID int, available bool) error {
	query := `UPDATE rooms SET is_available = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, available, roomID)
	if err != nil {
		return fmt.Errorf("failed to update room availability: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// HotelFilters структура для фильтрации отелей
type HotelFilters struct {
	City      string
	RoomType  string
	MaxGuests int
	MinPrice  float64
	MaxPrice  float64
}

// GetHotelsWithFilters возвращает отели с фильтрацией
func (r *HotelRepository) GetHotelsWithFilters(ctx context.Context, filters HotelFilters) ([]Hotel, error) {
	query := `
        SELECT DISTINCT h.id, h.name, h.address, h.contact_phone
        FROM hotels h
        LEFT JOIN room_types_in_hotels rth ON h.id = rth.hotel_id
        LEFT JOIN rooms r ON rth.id = r.room_type
        WHERE 1=1
    `

	var args []interface{}
	argIndex := 1

	if filters.City != "" {
		query += fmt.Sprintf(" AND h.address ILIKE $%d", argIndex)
		args = append(args, "%"+filters.City+"%")
		argIndex++
	}

	if filters.RoomType != "" {
		query += fmt.Sprintf(" AND rth.type = $%d", argIndex)
		args = append(args, filters.RoomType)
		argIndex++
	}

	if filters.MaxGuests > 0 {
		query += fmt.Sprintf(" AND rth.max_guests >= $%d", argIndex)
		args = append(args, filters.MaxGuests)
		argIndex++
	}

	if filters.MinPrice > 0 {
		query += fmt.Sprintf(" AND rth.price_per_night >= $%d", argIndex)
		args = append(args, filters.MinPrice)
		argIndex++
	}

	if filters.MaxPrice > 0 {
		query += fmt.Sprintf(" AND rth.price_per_night <= $%d", argIndex)
		args = append(args, filters.MaxPrice)
		argIndex++
	}

	query += " ORDER BY h.name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query hotels with filters: %w", err)
	}
	defer rows.Close()

	var hotels []Hotel
	for rows.Next() {
		var hotel Hotel
		err := rows.Scan(
			&hotel.ID,
			&hotel.Name,
			&hotel.Address,
			&hotel.ContactPhone,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hotel: %w", err)
		}
		hotels = append(hotels, hotel)
	}

	return hotels, nil
}

// CreateHotel creates new hotel
func (r *HotelRepository) CreateHotel(ctx context.Context, hotel *Hotel) (int, error) {
	query := `
		INSERT INTO hotels (name, address, contact_phone)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		hotel.Name,
		hotel.Address,
		hotel.ContactPhone,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create hotel: %w", err)
	}

	hotel.ID = id
	return id, nil
}

// UpdateHotel updates hotel information
func (r *HotelRepository) UpdateHotel(ctx context.Context, hotel *Hotel) error {
	query := `
		UPDATE hotels 
		SET name = $1, address = $2, contact_phone = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		hotel.Name,
		hotel.Address,
		hotel.ContactPhone,
		hotel.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update hotel: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// CreateRoom creates new room
func (r *HotelRepository) CreateRoom(ctx context.Context, room *Room) (int, error) {
	query := `
		INSERT INTO rooms (room_type, room_number, is_available)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		room.RoomType,
		room.RoomNumber,
		room.IsAvailable,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create room: %w", err)
	}

	room.ID = id
	return id, nil
}

// UpdateRoom updates room information
func (r *HotelRepository) UpdateRoom(ctx context.Context, room *Room) error {
	query := `
		UPDATE rooms 
		SET room_type = $1, room_number = $2, is_available = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		room.RoomType,
		room.RoomNumber,
		room.IsAvailable,
		room.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// GetRoomPriceInfo возвращает информацию о цене комнаты для gRPC
func (r *HotelRepository) GetRoomPriceInfo(ctx context.Context, hotelID, roomTypeID int) (float64, string, int, error) {
	query := `
        SELECT rth.price_per_night, 'RUB' as currency, r.id as room_id
        FROM room_types_in_hotels rth
        LEFT JOIN rooms r ON rth.id = r.room_type
        WHERE rth.hotel_id = $1 AND rth.id = $2 AND r.is_available = true
        LIMIT 1
    `

	var price float64
	var currency string
	var roomID int

	err := r.db.QueryRowContext(ctx, query, hotelID, roomTypeID).Scan(&price, &currency, &roomID)
	if err == sql.ErrNoRows {
		return 0, "", 0, ErrNotFound
	}
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to get room price: %w", err)
	}

	return price, currency, roomID, nil
}

// CalculateBookingPrice counts booking price
func (r *HotelRepository) CalculateBookingPrice(ctx context.Context, roomTypeID int, nights int, guests int) (float64, error) {
	roomType, err := r.GetRoomTypeByID(ctx, roomTypeID)
	if err != nil {
		return 0, err
	}

	if guests > roomType.MaxGuests {
		return 0, fmt.Errorf("maximum guests for this room is %d", roomType.MaxGuests)
	}

	return roomType.PricePerNight * float64(nights), nil
}
