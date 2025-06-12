package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cinema-reservation/internal/models"
	"cinema-reservation/internal/repositories"

	"github.com/gosimple/slug"
)

type cinemaService struct {
	cinemaRepo repositories.CinemaRepository
}

func NewCinemaService(cinemaRepo repositories.CinemaRepository) CinemaService {
	return &cinemaService{cinemaRepo: cinemaRepo}
}

func (s *cinemaService) CreateLayout(ctx context.Context, req *models.CreateCinemaRequest) (*models.Cinema, error) {
	// Trim and validate name
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("cinema name cannot be empty")
	}

	// Check if cinema name already exists
	exists, err := s.cinemaRepo.ExistsByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check cinema name existence: %w", err)
	}
	if exists {
		return nil, errors.New("cinema with this name already exists")
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
		return nil, fmt.Errorf("failed to create cinema: %w", err)
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
