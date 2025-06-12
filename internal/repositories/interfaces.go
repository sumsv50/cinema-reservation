package repositories

import (
	"context"

	"cinema-reservation/internal/models"
)

type CinemaRepository interface {
	Create(ctx context.Context, cinema *models.Cinema) error
	GetBySlug(ctx context.Context, slug string) (*models.Cinema, error)
	GetReservedSeats(ctx context.Context, cinemaID uint) ([]models.ReservedSeat, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

type ReservationRepository interface {
	Create(ctx context.Context, reservation *models.Reservation) error
	IsSeatsAvailable(ctx context.Context, cinemaID uint, seats []models.ReservedSeat) (bool, error)
	LockSeats(ctx context.Context, cinemaID uint, seats []models.ReservedSeat) error
	UnlockSeats(ctx context.Context, cinemaID uint, seats []models.ReservedSeat) error
}
