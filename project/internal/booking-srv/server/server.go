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

}

func (server *BookingServer) GetAllClientBookingsHandler(w http.ResponseWriter, r *http.Request) {

}

func (server *BookingServer) GetAllHotelBookingsHandler(w http.ResponseWriter, r *http.Request) {

}
