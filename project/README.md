# Система бронирования отелей

## Структура БД
- **booking_db**: users, bookings
- **hotel_db**: hotels, rooms

## Использование в коде
```go
import "project/internal/database"

// Подключение
bookingDB, _ := database.ConnectBookingDB(cfg)
hotelDB, _ := database.ConnectHotelDB(cfg)

// Репозитории
bookingRepo := database.NewBookingRepository(bookingDB)
hotelRepo := database.NewHotelRepository(hotelDB)
```