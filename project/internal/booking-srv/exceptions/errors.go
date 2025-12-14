package exceptions

import "errors"

var (
	ErrProblemsWithHotelManager = errors.New("there must be problems with hotel manager or your query is inaccurate")
	ErrDates                    = errors.New("wrong dates")
	ErrInsufficientFunds        = errors.New("insufficient funds")
	ErrNotFound                 = errors.New("not found")
)
