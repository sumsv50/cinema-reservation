package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"cinema-reservation/internal/models"
	"cinema-reservation/internal/repositories"
	"cinema-reservation/internal/utils"

	"github.com/go-redis/redis/v8"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
)

type cinemaService struct {
	cinemaRepo repositories.CinemaRepository
	redis      *redis.Client
}

func NewCinemaService(cinemaRepo repositories.CinemaRepository, redis *redis.Client) CinemaService {
	return &cinemaService{cinemaRepo: cinemaRepo, redis: redis}
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

func (s *cinemaService) GetAvailableSeats(ctx context.Context, slug string, groupSize int) ([][]models.Seat, error) {
	// Get cinema
	cinema, err := s.cinemaRepo.GetBySlug(ctx, slug)
	if err != nil {
		logrus.WithError(err).Error("failed to get cinema by slug")
		return nil, utils.ErrInternalServer
	}
	if cinema == nil {
		return nil, utils.ErrCinemaNotFound
	}

	reserved, err := s.GetRedisReservedSeats(ctx, cinema.ID)
	if err != nil {
		logrus.WithError(err).Error("failed to get reserved seats from redis")
		return nil, utils.ErrInternalServer
	}

	heatmap := buildHeatmap(cinema.Rows, cinema.Columns, cinema.MinDistance, reserved)
	available := FindSafeBlocks(heatmap, groupSize)

	return available, nil
}

func (s *cinemaService) GetRedisReservedSeats(ctx context.Context, cinemaID uint) ([]string, error) {
	key := fmt.Sprintf("cinema:%d:seats", cinemaID)

	data, err := s.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reserved seats from redis hash: %w", err)
	}

	var reserved []string
	for seat := range data {
		reserved = append(reserved, seat)
	}

	return reserved, nil
}

func buildHeatmap(rows, cols, minDist int, reserved []string) [][]bool {
	heat := make([][]bool, rows)
	for i := range heat {
		heat[i] = make([]bool, cols)
	}

	for _, seat := range reserved {
		parts := strings.Split(seat, ":")
		r, _ := strconv.Atoi(parts[0])
		c, _ := strconv.Atoi(parts[1])

		for dr := -minDist; dr <= minDist; dr++ {
			for dc := -minDist; dc <= minDist; dc++ {
				nr := r + dr
				nc := c + dc
				if nr >= 0 && nr < rows && nc >= 0 && nc < cols {
					if abs(dr)+abs(dc) < minDist {
						heat[nr][nc] = true // unsafe
					}
				}
			}
		}
	}
	return heat
}

func FindSafeBlocks(heat [][]bool, groupSize int) [][]models.Seat {
	rows := len(heat)
	cols := len(heat[0])
	var results [][]models.Seat

	for r := 0; r < rows; r++ {
		for c := 0; c <= cols-groupSize; c++ {
			valid := true
			block := []models.Seat{}

			for i := 0; i < groupSize; i++ {
				if heat[r][c+i] {
					valid = false
					break
				}
				block = append(block, models.Seat{Row: r, Column: c + i})
			}

			if valid {
				results = append(results, block)
			}
		}
	}

	return results
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
