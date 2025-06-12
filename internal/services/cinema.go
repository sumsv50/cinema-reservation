package services

import (
	"context"
	"fmt"
	"strings"

	"cinema-reservation/internal/models"
	"cinema-reservation/internal/repositories"
	"cinema-reservation/internal/utils"

	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
)

type cinemaService struct {
	cinemaRepo repositories.CinemaRepository
}

func NewCinemaService(cinemaRepo repositories.CinemaRepository) CinemaService {
	return &cinemaService{cinemaRepo: cinemaRepo}
}

func (s *cinemaService) CreateLayout(ctx context.Context, req *models.CreateCinemaRequest) (*models.Cinema, error) {
	// Trim name
	name := strings.TrimSpace(req.Name)

	// Check if cinema name already exists
	exists, err := s.cinemaRepo.ExistsByName(ctx, name)
	if err != nil {
		logrus.WithError(err).Error("failed to check cinema name existence")
		return nil, utils.ErrInternalServer
	}
	if exists {
		return nil, utils.ErrCinemaAlreadyExists
	}

	// Generate slug from name (since name is unique, slug will be unique too)
	cinemaSlug := slug.Make(name)

	cinema := &models.Cinema{
		Name:        name,
		Slug:        cinemaSlug,
		Rows:        req.Rows,
		Columns:     req.Columns,
		MinDistance: req.MinDistance,
	}

	err = s.cinemaRepo.Create(ctx, cinema)
	if err != nil {
		logrus.WithError(err).Error("failed to create cinema")
		return nil, utils.ErrInternalServer
	}

	return cinema, nil
}

func (s *cinemaService) GetAvailableSeats(ctx context.Context, slug string) (*models.CinemaLayout, error) {
	cinema, err := s.cinemaRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	reservedSeats, err := s.cinemaRepo.GetReservedSeats(ctx, cinema.ID)
	if err != nil {
		return nil, err
	}

	// Create seat map
	reservedMap := make(map[string]bool)
	for _, seat := range reservedSeats {
		key := fmt.Sprintf("%d-%d", seat.Row, seat.Column)
		reservedMap[key] = true
	}

	// Generate all seats
	var seats []models.Seat
	for row := 1; row <= cinema.Rows; row++ {
		for col := 1; col <= cinema.Columns; col++ {
			key := fmt.Sprintf("%d-%d", row, col)
			seats = append(seats, models.Seat{
				Row:      row,
				Column:   col,
				Reserved: reservedMap[key],
			})
		}
	}

	return &models.CinemaLayout{
		Cinema: *cinema,
		Seats:  seats,
	}, nil
}
