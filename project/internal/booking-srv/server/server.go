package server

import (
	"encoding/json"
	"net/http"

	"hotel-booking-system/internal/booking-srv/stg"
	api "hotel-booking-system/package/api/stable"
)

type BookingServer struct {
	Src *stg.Storage
	Mux *http.ServeMux
}

func NewBookingServer(service *stg.Storage) *BookingServer {
	return &BookingServer{
		Src: service,
		Mux: http.NewServeMux(),
	}
}

func (server *BookingServer) SetServer() {
	server.Mux.HandleFunc("POST /api/create_booking", server.CreateBookingHandler)
	server.Mux.HandleFunc("GET /api/get_all_client_bookings", server.GetAllClientBookingsHandler)

	server.Mux.HandleFunc("GET /live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (server *BookingServer) CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var req api.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidJSON(w, http.StatusBadRequest)
		return
	}

	bookingInfo := stg.BookingInfo{
		UserID:       req.UserID,
		HotelID:      req.HotelID,
		RoomTypeID:   req.RoomTypeID,
		CheckInDate:  req.CheckInDate,
		CheckOutDate: req.CheckOutDate,
		GuestsCount:  req.GuestsCount,
	}

	bookingId, err := server.Src.CreateBooking(r.Context(), bookingInfo)
	if err != nil {
		writeInvalidJSONError(w, err)
		return
	}

	response := map[string]int{
		"booking_id": bookingId,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (server *BookingServer) GetAllClientBookingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var userID int
	if err := json.NewDecoder(r.Body).Decode(&userID); err != nil {
		writeInvalidJSON(w, http.StatusBadRequest)
		return
	}

	bookings, err := server.Src.GetAllClientBookings(r.Context(), userID)
	if err != nil {
		writeInvalidJSONError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(bookings)
}

func writeInvalidJSON(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: "invalid json",
	})
}

func writeInvalidJSONError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
