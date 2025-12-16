package exceptions

import "errors"

var (
    ErrInvalidHotelData  = errors.New("invalid hotel data")
    ErrInvalidRoomData   = errors.New("invalid room data")
    ErrHotelNotFound     = errors.New("hotel not found")
    ErrRoomNotFound      = errors.New("room not found")
    ErrRoomTypeNotFound  = errors.New("room type not found")
    ErrRoomNotAvailable  = errors.New("room not available")
    ErrInvalidPrice      = errors.New("invalid price")
)