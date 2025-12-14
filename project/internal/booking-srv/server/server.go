package server

import (
	"encoding/json"
	"hotel-booking-system/internal/booking-srv/stg"
	"net/http"
)

func writeInvalidJSON(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(err{Error: "invalid json"})
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

type BookingServer struct {
	Src *stg.Storage
	Mux *http.ServeMux
}

type err struct {
	Error string `json:"error"`
}

func NewBookingServer(service *stg.Storage) *BookingServer {
	return &BookingServer{Src: service, Mux: http.NewServeMux()}
}

func (server *BookingServer) SetServer() {
	server.Mux.HandleFunc("POST /api/create_booking", server.CreateBookingHandler)
	server.Mux.HandleFunc("GET /api/get_all_client_bookings", server.GetAllClientBookingsHandler)
	server.Mux.HandleFunc("GET /api/get_all_hotel_bookings", server.GetAllHotelBookingsHandler)
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

	request := stg.BookingInfo{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeInvalidJSON(w, http.StatusBadRequest)
		return
	}

	bookingId, err := server.Src.CreateBooking(request)

	if err != nil {
		writeInvalidJSONError(w, err)
		return
	}

	response := map[string]int{
		"booking_id": bookingId,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (server *BookingServer) GetAllClientBookingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	request := 0

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeInvalidJSON(w, http.StatusBadRequest)
		return
	}

	bookings, err := server.Src.GetAllClientBookings(request)

	if err != nil {
		writeInvalidJSONError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookings)
}

func (server *BookingServer) GetAllHotelBookingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	request := 0

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeInvalidJSON(w, http.StatusBadRequest)
		return
	}

	bookings, err := server.Src.GetAllHotelBookings(request)

	if err != nil {
		writeInvalidJSONError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookings)
}
