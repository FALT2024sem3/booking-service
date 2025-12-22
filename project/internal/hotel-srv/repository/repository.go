package repository

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("record not found")

type Hotel struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	ContactPhone string `json:"contact_phone"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateHotel(ctx context.Context, hotel *Hotel) error {
	query := `INSERT INTO hotels (name, address, contact_phone) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRowContext(ctx, query, hotel.Name, hotel.Address, hotel.ContactPhone).Scan(&hotel.ID)
}

func (r *Repository) GetAllHotels(ctx context.Context) ([]Hotel, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, address, contact_phone FROM hotels`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotels []Hotel
	for rows.Next() {
		var h Hotel
		if err := rows.Scan(&h.ID, &h.Name, &h.Address, &h.ContactPhone); err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}
	return hotels, nil
}

func (r *Repository) GetRoomPriceInfo(ctx context.Context, hotelID, roomTypeID int) (float64, string, error) {
	query := `
		SELECT price_per_night, 'RUB' 
		FROM room_types_in_hotels 
		WHERE hotel_id = $1 AND id = $2
	`
	var price float64
	var currency string
	err := r.db.QueryRowContext(ctx, query, hotelID, roomTypeID).Scan(&price, &currency)
	if err == sql.ErrNoRows {
		return 0, "", ErrNotFound
	}
	return price, currency, err
}

func (r *Repository) GetRoomIDsByHotelAndType(ctx context.Context, hotelID, roomTypeID int) ([]int, error) {

	query := `SELECT id FROM rooms WHERE room_type = $1`
	rows, err := r.db.QueryContext(ctx, query, roomTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
