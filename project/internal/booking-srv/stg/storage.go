package stg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hotel-booking-system/internal/booking-srv/repository"
	"hotel-booking-system/internal/kafka"
	"hotel-booking-system/package/events"
	hotelv1 "hotel-booking-system/package/proto/fast/stable"

	"github.com/sirupsen/logrus"
)

type BookingInfo struct {
	UserID       int       `json:"user_id"`
	HotelID      int       `json:"hotel_id"`
	RoomTypeID   int       `json:"room_type_id"`
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
	GuestsCount  int       `json:"guests_count"`
	UserEmail    string    `json:"user_email"`
	UserName     string    `json:"user_name"`
}

type Storage struct {
	repo        *repository.Repository
	hotelClient hotelv1.HotelServiceClient
	producer    *kafka.Producer
}

func NewStorage(repo *repository.Repository, client hotelv1.HotelServiceClient, producer *kafka.Producer) *Storage {
	return &Storage{
		repo:        repo,
		hotelClient: client,
		producer:    producer,
	}
}

func (s *Storage) CreateBooking(ctx context.Context, info BookingInfo) (int, error) {
	priceReq := &hotelv1.GetRoomPriceRequest{
		HotelId:    int32(info.HotelID),
		RoomTypeId: int32(info.RoomTypeID),
	}

	priceResp, err := s.hotelClient.GetRoomPrice(ctx, priceReq)
	if err != nil {
		logrus.Errorf("Failed to get room price: %v", err)
		return 0, fmt.Errorf("failed to get room price: %w", err)
	}

	days := int(info.CheckOutDate.Sub(info.CheckInDate).Hours() / 24)
	if days <= 0 {
		return 0, fmt.Errorf("invalid dates: check-out must be after check-in")
	}
	totalPrice := priceResp.Price * float64(days)

	roomsReq := &hotelv1.GetRoomsIDRequest{
		HotelId:    int32(info.HotelID),
		RoomTypeId: int32(info.RoomTypeID),
	}

	roomsResp, err := s.hotelClient.GetRoomsID(ctx, roomsReq)
	if err != nil {
		logrus.Errorf("Failed to get rooms list: %v", err)
		return 0, fmt.Errorf("failed to fetch rooms from hotel service: %w", err)
	}

	if len(roomsResp.RoomIds) == 0 {
		return 0, fmt.Errorf("no rooms found for this type in hotel")
	}

	busyRooms, err := s.repo.GetBusyRooms(ctx, info.CheckInDate, info.CheckOutDate)
	if err != nil {
		return 0, fmt.Errorf("failed to check local availability: %w", err)
	}

	var availableRoomID int

	for _, roomID := range roomsResp.RoomIds {
		if !busyRooms[int(roomID)] {
			availableRoomID = int(roomID)
			break
		}
	}

	if availableRoomID == 0 {
		return 0, fmt.Errorf("no available rooms for selected dates")
	}

	booking := &repository.Booking{
		UserID:       info.UserID,
		HotelID:      info.HotelID,
		RoomID:       availableRoomID,
		CheckInDate:  info.CheckInDate,
		CheckOutDate: info.CheckOutDate,
		GuestsCount:  info.GuestsCount,
		TotalPrice:   totalPrice,
	}

	bookingID, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		return 0, fmt.Errorf("failed to create booking in db: %w", err)
	}

	event := events.BookingCreatedEvent{
		BookingID:    bookingID,
		UserEmail:    info.UserEmail,
		UserName:     info.UserName,
		Amount:       totalPrice,
		CheckInDate:  info.CheckInDate.Format("2006-01-02"),
		CheckOutDate: info.CheckOutDate.Format("2006-01-02"),
	}

	payload, err := json.Marshal(event)
	if err == nil {
		if err := s.producer.Produce(string(payload), "booking-created"); err != nil {
			logrus.Errorf("Failed to send kafka event: %v", err)
		} else {
			logrus.Infof("Event sent to Kafka: %s", string(payload))
		}
	} else {
		logrus.Errorf("Failed to marshal kafka event: %v", err)
	}

	return bookingID, nil
}

func (s *Storage) GetAllClientBookings(ctx context.Context, userID int) ([]repository.Booking, error) {
	return s.repo.GetUserBookings(ctx, userID)
}

func (s *Storage) GetAllHotelBookings(ctx context.Context, hotelID int) ([]repository.Booking, error) {
	return s.repo.GetHotelBookings(ctx, hotelID)
}
