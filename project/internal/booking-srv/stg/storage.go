package stg

import (
	"context"
	"database/sql"
	"hotel-booking-system/internal/booking-srv/exceptions"
	"hotel-booking-system/internal/database"
	"hotel-booking-system/internal/kafka"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var address = []string{"localhost:9091", "localhost:9092", "localhost:9093"}

const (
	topic = "my-topic"
)

type BookingInfo struct {
	UserID       int       `json:"user_id"`
	HotelID      int       `json:"hotel_id"`
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
	GuestsCount  int       `json:"guests_count"`
	RoomType     int       `json:"room_type"`
}

type Storage struct {
	database       database.BookingRepository
	userRepository database.UserRepository
	mux            sync.RWMutex
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		database:       *database.NewBookingRepository(db),
		userRepository: *database.NewUserRepository(db),
	}
}

func DaysBetween(checkIn, checkOut time.Time) int {
	checkInStart := time.Date(checkIn.Year(), checkIn.Month(), checkIn.Day(), 0, 0, 0, 0, checkIn.Location())
	checkOutStart := time.Date(checkOut.Year(), checkOut.Month(), checkOut.Day(), 0, 0, 0, 0, checkOut.Location())

	duration := checkOutStart.Sub(checkInStart)
	days := int(duration.Hours()/24) + 1
	return days
}

func (storage *Storage) CreateBooking(bookingInfo BookingInfo) (int, error) {
	storage.mux.RLock()
	defer storage.mux.RUnlock()

	if bookingInfo.CheckOutDate.Before(bookingInfo.CheckInDate) {
		return 0, exceptions.ErrDates
	}

	// Andrey's Talalaev func
	hotelPricePerNight, roomId := float64(15), 2 //, exceptions.ErrProblemsWithHotelManager

	totalPrice := hotelPricePerNight * float64(DaysBetween(bookingInfo.CheckInDate, bookingInfo.CheckOutDate))

	booking := database.Booking{
		UserID:       bookingInfo.UserID,
		HotelID:      bookingInfo.HotelID,
		RoomID:       roomId,
		CheckInDate:  bookingInfo.CheckInDate,
		CheckOutDate: bookingInfo.CheckOutDate,
		GuestsCount:  bookingInfo.GuestsCount,
		TotalPrice:   float64(totalPrice),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bookingID, err := storage.database.CreateBooking(ctx, &booking)
	if err != nil {
		return 0, err
	}

	user, err := storage.userRepository.GetUserByID(ctx, bookingInfo.UserID)
	producer, err := kafka.NewProducer(address)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Printf("Starting producer for %s...", user.Email)
	if err = producer.Produce(user.Email, topic); err != nil {
		logrus.Error(err)
	}
	logrus.Printf("Finished producing messages for %s", user.Email)
	producer.Close()
	return bookingID, nil
}

func (storage *Storage) GetAllClientBookings(userId int) ([]database.Booking, error) {
	storage.mux.RLock()
	defer storage.mux.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bookings, err := storage.database.GetUserBookings(ctx, userId)
	if err != nil {
		return []database.Booking{}, err
	}

	return bookings, nil
}

func (storage *Storage) GetAllHotelBookings(hotelId int) ([]database.Booking, error) {
	storage.mux.RLock()
	defer storage.mux.RUnlock()

	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	// Artem's Dmitroc func
	//bookings, err := storage.database.GetHotelBookings(ctx, hotelId)
	//if err != nil {
	//	return []database.Booking{}, err
	//}

	return nil, nil
}
