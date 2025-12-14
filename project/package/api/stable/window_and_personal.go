package stable

import "time"

type CreateBookingRequest struct {
	UserID       int       `json:"user_id"`
	HotelID      int       `json:"hotel_id"`
	RoomTypeID   int       `json:"room_type_id"`
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
	GuestsCount  int       `json:"guests_count"`
}

type CreateBookingResponse struct {
	BookingID int    `json:"booking_id"`
	Status    string `json:"status"`
}

type GetUserBookingsResponse struct {
	Bookings []BookingDTO `json:"bookings"`
}

type BookingDTO struct {
	ID          int       `json:"id"`
	HotelName   string    `json:"hotel_name"`
	CheckInDate time.Time `json:"check_in_date"`
	TotalPrice  float64   `json:"total_price"`
}
