package services

import (
	"context"
	"fmt"
	"strings"

	"cinema-reservation/internal/models"
	"cinema-reservation/internal/repositories"
	scriptloader "cinema-reservation/internal/scripts"
	"cinema-reservation/internal/utils"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type reservationService struct {
	reservationRepo repositories.ReservationRepository
	cinemaRepo      repositories.CinemaRepository
	redis           *redis.Client
}

func NewReservationService(
	reservationRepo repositories.ReservationRepository,
	cinemaRepo repositories.CinemaRepository,
	redis *redis.Client,
) ReservationService {
	return &reservationService{
		reservationRepo: reservationRepo,
		cinemaRepo:      cinemaRepo,
		redis:           redis,
	}
}

func (s *reservationService) ReserveSeats(ctx context.Context, req *models.ReservationRequest) (*models.Reservation, error) {
	// Get cinema
	cinema, err := s.cinemaRepo.GetBySlug(ctx, req.CinemaSlug)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	if cinema == nil {
		return nil, utils.ErrCinemaNotFound
	}

	// Validate seats
	var reservedSeats []models.ReservedSeat
	for _, seat := range req.Seats {
		if seat.Row < 0 || seat.Row >= cinema.Rows || seat.Column < 0 || seat.Column >= cinema.Columns {
			return nil, utils.ErrInvalidSeatPosition
		}
		reservedSeats = append(reservedSeats, models.ReservedSeat{
			CinemaID: cinema.ID,
			Row:      seat.Row,
			Column:   seat.Column,
		})
	}

	err = s.reserveSeatsRedis(ctx, cinema.ID, reservedSeats, cinema.MinDistance)
	if err != nil {
		return nil, err
	}

	// Create reservation
	reservation := &models.Reservation{
		CinemaID: cinema.ID,
		Note:     req.Note,
		Seats:    reservedSeats,
	}

	err = s.reservationRepo.Create(ctx, reservation)
	if err != nil {
		cancelErr := s.rollbackSeatsRedis(ctx, cinema.ID, reservedSeats)
		if cancelErr != nil {
			logrus.WithFields(logrus.Fields{
				"cinema_id":      cinema.ID,
				"reserved_seats": models.ReservedSeats(reservedSeats).String(),
				"rollback_error": cancelErr.Error(),
				"original_error": err.Error(),
				"operation":      "seat_reservation_rollback",
			}).Error("CRITICAL: Failed to rollback reserved seats on Redis after reservation creation failed - manual intervention required")
		}

		// TODO: Add retry mechanism and send notification to admin if totally failed

		logrus.WithError(err).Error("insert reservation to DB failed")
		return nil, utils.ErrInternalServer
	}

	return reservation, nil
}

func (s *reservationService) reserveSeatsRedis(ctx context.Context, cinemaID uint, seats []models.ReservedSeat, minDist int) error {
	script, err := scriptloader.LoadReserveScript()
	if err != nil {
		logrus.WithError(err).Error("load script failed")
		return utils.ErrInternalServer
	}

	args := []interface{}{minDist}
	for _, s := range seats {
		args = append(args, fmt.Sprintf("%d:%d", s.Row, s.Column))
	}
	key := fmt.Sprintf("cinema:%d:seats", cinemaID)

	result, err := script.Run(ctx, s.redis, []string{key}, args...).Result()
	if err != nil {
		logrus.WithError(err).Error("seat reservation failed")

		if strings.HasPrefix(err.Error(), "[SEATS_RESERVED]") {
			return utils.ErrSeatsAlreadyReserved
		} else if strings.HasPrefix(err.Error(), "[MIN_DISTANCE_VIOLATION]") {
			return utils.ErrMinDistanceViolation
		}

		return utils.ErrInternalServer
	}

	if result != "OK" {
		logrus.Errorf("unexpected result: %v", result)
		return utils.ErrInternalServer
	}

	return nil
}

func (s *reservationService) rollbackSeatsRedis(ctx context.Context, cinemaID uint, seats []models.ReservedSeat) error {
	script, err := scriptloader.LoadCancelScript()
	if err != nil {
		return fmt.Errorf("load script failed: %w", err)
	}

	args := []interface{}{}
	for _, s := range seats {
		args = append(args, fmt.Sprintf("%d:%d", s.Row, s.Column))
	}
	key := fmt.Sprintf("cinema:%d:seats", cinemaID)

	result, err := script.Run(ctx, s.redis, []string{key}, args...).Result()
	if err != nil {
		return fmt.Errorf("cancel seats failed: %w", err)
	}

	if result != "OK" {
		return fmt.Errorf("unexpected result: %v", result)
	}

	return nil
}
