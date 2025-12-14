package events

type BookingCreatedEvent struct {
	BookingID    int     `json:"booking_id"`
	UserEmail    string  `json:"user_email"`
	UserName     string  `json:"user_name"`
	HotelName    string  `json:"hotel_name"`
	CheckInDate  string  `json:"check_in_date"`
	CheckOutDate string  `json:"check_out_date"`
	Amount       float64 `json:"amount"`
}
