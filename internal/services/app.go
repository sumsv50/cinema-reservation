package services

import (
	"cinema-reservation/internal/repositories"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type appService struct {
	reservationRepo repositories.ReservationRepository
	redis           *redis.Client
}

func NewAppService(reservationRepo repositories.ReservationRepository, redis *redis.Client) AppService {
	return &appService{reservationRepo: reservationRepo, redis: redis}
}

func (s *appService) SyncReservationsToRedis() error {
	ctx := context.Background()
	logrus.Println("Starting simple sync of reservations to Redis...")
	startTime := time.Now()

	// Get all reserved seats
	reservedSeats, err := s.reservationRepo.GetAllReservedSeats(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch reserved seats: %w", err)
	}

	logrus.Printf("Found %d reserved seats to sync", len(reservedSeats))

	cinemaSeats := make(map[uint][]string)
	for _, seat := range reservedSeats {
		seatKey := fmt.Sprintf("%d:%d", seat.Row, seat.Column)
		cinemaSeats[seat.CinemaID] = append(cinemaSeats[seat.CinemaID], seatKey)
	}

	pipe := s.redis.Pipeline()

	keys, err := s.redis.Keys(ctx, "cinema:*:seats").Result()
	if err == nil && len(keys) > 0 {
		pipe.Del(ctx, keys...)
	}

	// Set reserved seats for each cinema
	for cinemaID, seatKeys := range cinemaSeats {
		key := fmt.Sprintf("cinema:%d:seats", cinemaID)

		args := make([]interface{}, 0, len(seatKeys)*2)
		for _, seatKey := range seatKeys {
			args = append(args, seatKey, "1")
		}

		if len(args) > 0 {
			pipe.HSet(ctx, key, args...)
		}
	}

	// Execute all operations
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	duration := time.Since(startTime)
	logrus.Printf("Successfully synced %d reserved seats across %d cinemas to Redis in %v",
		len(reservedSeats), len(cinemaSeats), duration)

	return nil
}
