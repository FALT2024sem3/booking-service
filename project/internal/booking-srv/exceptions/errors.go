package exceptions

import "errors"

var (
	ErrInvalidAmount     = errors.New("invalid amount: must be positive")
	ErrSelfTransfer      = errors.New("self transfer is not allowed")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrNotFound          = errors.New("not found")
)
