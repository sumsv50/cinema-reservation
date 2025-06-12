package services

import (
	"context"

	"cinema-reservation/internal/models"
)

type CinemaService interface {
	CreateLayout(ctx context.Context, req *models.CreateCinemaRequest) (*models.Cinema, error)
	GetAvailableSeats(ctx context.Context, slug string) (*models.CinemaLayout, error)
}

type ReservationService interface {
	ReserveSeats(ctx context.Context, req *models.ReservationRequest) (*models.Reservation, error)
	CancelSeats(ctx context.Context, req *models.CancelRequest) error
}
