package stg

import (
	"context"
	"database/sql"
	"hotel-booking-system/internal/database"
	"sync"
)

type Storage struct {
	database *database.HotelRepository
	mux      sync.RWMutex
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		database: database.NewHotelRepository(db),
	}
}

// GetHotelsWithFilters возвращает отели с фильтрацией
func (storage *Storage) GetHotelsWithFilters(ctx context.Context, filters database.HotelFilters) ([]database.Hotel, error) {
	storage.mux.RLock()
	defer storage.mux.RUnlock()

	hotels, err := storage.database.GetHotelsWithFilters(ctx, filters)
	if err != nil {
		return nil, err
	}

	return hotels, nil
}

// CreateOrUpdateHotel создает или обновляет отель
func (storage *Storage) CreateOrUpdateHotel(ctx context.Context, hotel *database.Hotel) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	if hotel.ID == 0 {
		// Создание нового отеля
		_, err := storage.database.CreateHotel(ctx, hotel)
		return err
	} else {
		// Обновление существующего отеля
		return storage.database.UpdateHotel(ctx, hotel)
	}
}

// CreateOrUpdateRoom создает или обновляет комнату
func (storage *Storage) CreateOrUpdateRoom(ctx context.Context, room *database.Room) error {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	if room.ID == 0 {
		// Создание новой комнаты
		_, err := storage.database.CreateRoom(ctx, room)
		return err
	} else {
		// Обновление существующей комнаты
		return storage.database.UpdateRoom(ctx, room)
	}
}

// GetRoomPriceInfo возвращает информацию о цене комнаты
func (storage *Storage) GetRoomPriceInfo(ctx context.Context, hotelID, roomTypeID int) (float64, string, int, error) {
	storage.mux.RLock()
	defer storage.mux.RUnlock()

	return storage.database.GetRoomPriceInfo(ctx, hotelID, roomTypeID)
}
