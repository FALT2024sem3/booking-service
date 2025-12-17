package stg

import (
	"context"
	"hotel-booking-system/internal/hotel-srv/repository"
)

type Storage struct {
	repo *repository.Repository
}

func NewStorage(repo *repository.Repository) *Storage {
	return &Storage{
		repo: repo,
	}
}

func (s *Storage) CreateHotel(ctx context.Context, hotel *repository.Hotel) error {
	return s.repo.CreateHotel(ctx, hotel)
}

func (s *Storage) GetAllHotels(ctx context.Context) ([]repository.Hotel, error) {
	return s.repo.GetAllHotels(ctx)
}

func (s *Storage) GetRoomPriceInfo(ctx context.Context, hotelID, roomTypeID int) (float64, string, error) {
	return s.repo.GetRoomPriceInfo(ctx, hotelID, roomTypeID)
}

func (s *Storage) GetRoomIDsByHotelAndType(ctx context.Context, hotelID, roomTypeID int) ([]int, error) {
	return s.repo.GetRoomIDsByHotelAndType(ctx, hotelID, roomTypeID)
}
