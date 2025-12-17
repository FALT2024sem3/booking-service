package server

import (
	"context"
	"encoding/json"
	"net/http"

	"hotel-booking-system/internal/hotel-srv/repository"
	"hotel-booking-system/internal/hotel-srv/stg"
	hotelv1 "hotel-booking-system/package/proto/fast/stable"

	"github.com/sirupsen/logrus"
)

type HotelServer struct {
	Src *stg.Storage
	Mux *http.ServeMux
	hotelv1.UnimplementedHotelServiceServer
}

func NewHotelServer(storage *stg.Storage) *HotelServer {
	return &HotelServer{
		Src: storage,
		Mux: http.NewServeMux(),
	}
}

func (server *HotelServer) SetServer() {
	server.Mux.HandleFunc("GET /api/hotels", server.GetHotelsHandler)
	server.Mux.HandleFunc("POST /api/hotels", server.CreateHotelHandler)

	server.Mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func writeInvalidJSON(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: "invalid json",
	})
}

func writeServerError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
	logrus.Error(err)
}

func (server *HotelServer) GetHotelsHandler(w http.ResponseWriter, r *http.Request) {
	logrus.Info("GetHotels request")

	hotels, err := server.Src.GetAllHotels(r.Context())
	if err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hotels)
}

func (server *HotelServer) CreateHotelHandler(w http.ResponseWriter, r *http.Request) {
	logrus.Info("CreateHotel request")

	var hotel repository.Hotel

	if err := json.NewDecoder(r.Body).Decode(&hotel); err != nil {
		writeInvalidJSON(w, http.StatusBadRequest)
		return
	}

	if hotel.Name == "" || hotel.Address == "" || hotel.ContactPhone == "" {
		http.Error(w, "name, address and contact_phone are required", http.StatusBadRequest)
		return
	}

	if err := server.Src.CreateHotel(r.Context(), &hotel); err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hotel)
}

func (server *HotelServer) GetRoomPrice(ctx context.Context, req *hotelv1.GetRoomPriceRequest) (*hotelv1.GetRoomPriceResponse, error) {
	logrus.WithFields(logrus.Fields{
		"hotel_id":     req.HotelId,
		"room_type_id": req.RoomTypeId,
	}).Info("GetRoomPrice gRPC request")

	price, currency, err := server.Src.GetRoomPriceInfo(ctx, int(req.HotelId), int(req.RoomTypeId))
	if err != nil {
		logrus.WithError(err).Error("Failed to get room price")
		return nil, err
	}

	return &hotelv1.GetRoomPriceResponse{
		Price:    price,
		Currency: currency,
	}, nil
}

func (server *HotelServer) GetRoomsID(ctx context.Context, req *hotelv1.GetRoomsIDRequest) (*hotelv1.GetRoomsIDResponse, error) {
	logrus.WithFields(logrus.Fields{
		"hotel_id":     req.HotelId,
		"room_type_id": req.RoomTypeId,
	}).Info("GetRoomsID gRPC request")

	ids, err := server.Src.GetRoomIDsByHotelAndType(ctx, int(req.HotelId), int(req.RoomTypeId))
	if err != nil {
		logrus.WithError(err).Error("Failed to get room IDs")
		return nil, err
	}

	var protoIDs []int32
	for _, id := range ids {
		protoIDs = append(protoIDs, int32(id))
	}

	return &hotelv1.GetRoomsIDResponse{
		RoomIds: protoIDs,
	}, nil
}
