package server

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    
    "hotel-booking-system/internal/database"
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
    server.Mux.HandleFunc("POST /api/hotels", server.CreateOrUpdateHotelHandler)
    server.Mux.HandleFunc("POST /api/rooms", server.CreateOrUpdateRoomHandler)
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

func writeInvalidJSONError(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    
    json.NewEncoder(w).Encode(struct {
        Error string `json:"error"`
    }{
        Error: err.Error(),
    })
}

func writeServerError(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    
    json.NewEncoder(w).Encode(struct {
        Error string `json:"error"`
    }{
        Error: "internal server error",
    })
    logrus.Error(err)
}

// GetHotelsHandler возвращает список отелей с фильтрацией
func (server *HotelServer) GetHotelsHandler(w http.ResponseWriter, r *http.Request) {
    logrus.WithFields(logrus.Fields{
        "method": r.Method,
        "path":   r.URL.Path,
        "query":  r.URL.Query(),
    }).Info("GetHotels request")
    
    // Парсинг параметров фильтрации
    filters := database.HotelFilters{
        City:      r.URL.Query().Get("city"),
        RoomType:  r.URL.Query().Get("room_type"),
    }
    
    if maxGuests := r.URL.Query().Get("max_guests"); maxGuests != "" {
        if val, err := strconv.Atoi(maxGuests); err == nil {
            filters.MaxGuests = val
        }
    }
    
    if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
        if val, err := strconv.ParseFloat(minPrice, 64); err == nil {
            filters.MinPrice = val
        }
    }
    
    if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
        if val, err := strconv.ParseFloat(maxPrice, 64); err == nil {
            filters.MaxPrice = val
        }
    }
    
    // Получение отелей
    hotels, err := server.Src.GetHotelsWithFilters(r.Context(), filters)
    if err != nil {
        writeServerError(w, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(hotels)
}

// CreateOrUpdateHotelHandler создает или обновляет отель
func (server *HotelServer) CreateOrUpdateHotelHandler(w http.ResponseWriter, r *http.Request) {
    logrus.WithFields(logrus.Fields{
        "method": r.Method,
        "path":   r.URL.Path,
    }).Info("CreateOrUpdateHotel request")
    
    var hotel database.Hotel
    
    if err := json.NewDecoder(r.Body).Decode(&hotel); err != nil {
        writeInvalidJSON(w, http.StatusBadRequest)
        return
    }
    
    // Валидация
    if hotel.Name == "" || hotel.Address == "" || hotel.ContactPhone == "" {
        writeInvalidJSONError(w, fmt.Errorf("name, address and contact_phone are required"))
        return
    }
    
    if err := server.Src.CreateOrUpdateHotel(r.Context(), &hotel); err != nil {
        writeServerError(w, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(hotel)
}

// CreateOrUpdateRoomHandler создает или обновляет комнату
func (server *HotelServer) CreateOrUpdateRoomHandler(w http.ResponseWriter, r *http.Request) {
    logrus.WithFields(logrus.Fields{
        "method": r.Method,
        "path":   r.URL.Path,
    }).Info("CreateOrUpdateRoom request")
    
    var room database.Room
    
    if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
        writeInvalidJSON(w, http.StatusBadRequest)
        return
    }
    
    // Валидация
    if room.RoomNumber == "" || room.RoomType == 0 {
        writeInvalidJSONError(w, fmt.Errorf("room_number and room_type are required"))
        return
    }
    
    if err := server.Src.CreateOrUpdateRoom(r.Context(), &room); err != nil {
        writeServerError(w, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(room)
}

// GetRoomPrice - реализация gRPC метода
func (server *HotelServer) GetRoomPrice(ctx context.Context, req *hotelv1.GetRoomPriceRequest) (*hotelv1.GetRoomPriceResponse, error) {
    logrus.WithFields(logrus.Fields{
        "hotel_id":      req.HotelId,
        "room_type_id":  req.RoomTypeId,
    }).Info("GetRoomPrice gRPC request")
    
    price, currency, roomID, err := server.Src.GetRoomPriceInfo(ctx, int(req.HotelId), int(req.RoomTypeId))
    if err != nil {
        logrus.WithError(err).Error("Failed to get room price")
        return nil, err
    }
    
    return &hotelv1.GetRoomPriceResponse{
        Price:    price,
        Currency: currency,
        RoomId:   int32(roomID),
    }, nil
}