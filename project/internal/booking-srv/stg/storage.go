package stg

import (
	"database/sql"
	"hotel-booking-system/internal/database"
	"sync"
	"time"
)

type BookingInfo struct {
	UserID       int
	HotelID      int
	CheckInDate  time.Time
	CheckOutDate time.Time
	GuestsCount  int
	RoomType     int
}

type Storage struct {
	database database.BookingRepository
	mux      sync.RWMutex
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		database: *database.NewBookingRepository(db),
	}
}

func (storage *Storage) CreateBooking(bookingInfo BookingInfo) (int, error) {
	storage.mux.RLock()
	defer storage.mux.RUnlock()
}

func (storage *Storage) GetAllClientBookings(userId int) ([]BookingInfo, error) {
	storage.mux.RLock()
	defer storage.mux.RUnlock()
}

func (storage *Storage) GetAllHotelBookings(hotelId int) ([]BookingInfo, error) {
	storage.mux.RLock()
	defer storage.mux.RUnlock()
}
