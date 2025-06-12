package utils

import "errors"

var (
	ErrCinemaNotFound       = errors.New("cinema not found")
	ErrCinemaAlreadyExists  = errors.New("cinema with this name already exists")
	ErrInvalidInput         = errors.New("invalid input provided")
	ErrSeatsAlreadyReserved = errors.New("one or more seats are already reserved")
	ErrSeatsNotAvailable    = errors.New("selected seats are not available")
	ErrInvalidSeatPosition  = errors.New("invalid seat position")
	ErrMinDistanceViolation = errors.New("seat selection violates minimum distance requirement")
	ErrInternalServer       = errors.New("internal server error")
	ErrDatabaseConnection   = errors.New("database connection failed")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrSeatsNotReserved     = errors.New("one or more seats are not currently reserved")
)
