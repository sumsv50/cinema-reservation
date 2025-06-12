package repositories

import (
	"context"
	"fmt"
	"log"

	"cinema-reservation/internal/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type reservationRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewReservationRepository(db *gorm.DB, redis *redis.Client) ReservationRepository {
	return &reservationRepository{db: db, redis: redis}
}

func (r *reservationRepository) Create(ctx context.Context, reservation *models.Reservation) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(reservation).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *reservationRepository) IsSeatsAvailable(ctx context.Context, cinemaID uint, seats []models.ReservedSeat) (bool, error) {
	for _, seat := range seats {
		// Check in database
		var count int64
		err := r.db.WithContext(ctx).Model(&models.ReservedSeat{}).
			Where("cinema_id = ? AND row = ? AND column = ?", cinemaID, seat.Row, seat.Column).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		if count > 0 {
			return false, nil
		}

		// Check in Redis (temporary locks)
		key := fmt.Sprintf("seat_lock:%d:%d:%d", cinemaID, seat.Row, seat.Column)
		exists, err := r.redis.Exists(ctx, key).Result()
		if err != nil {
			return false, err
		}
		if exists > 0 {
			return false, nil
		}
	}
	return true, nil
}

func (r *reservationRepository) FindReservedSeats(ctx context.Context, cinemaID uint, seats []models.Seat) ([]models.ReservedSeat, error) {
	if len(seats) == 0 {
		return []models.ReservedSeat{}, nil
	}

	var reservedSeats []models.ReservedSeat

	query := r.db.WithContext(ctx).Where("cinema_id = ?", cinemaID)

	var orConditions []string
	var args []interface{}

	for _, seat := range seats {
		orConditions = append(orConditions, `("row" = ? AND "column" = ?)`)
		args = append(args, seat.Row, seat.Column)
	}

	if len(orConditions) > 0 {
		whereClause := "(" + orConditions[0]
		for i := 1; i < len(orConditions); i++ {
			whereClause += " OR " + orConditions[i]
		}
		whereClause += ")"

		query = query.Where(whereClause, args...)
	}

	err := query.Find(&reservedSeats).Error
	if err != nil {
		log.Printf("Error finding reserved seats: %v", err)
		return nil, err
	}

	log.Printf("Found %d reserved seats out of %d requested", len(reservedSeats), len(seats))
	return reservedSeats, nil
}

func (r *reservationRepository) CancelSeats(ctx context.Context, seatIDs []uint) error {
	if len(seatIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Soft delete all seats in one operation
		result := tx.Delete(&models.ReservedSeat{}, seatIDs)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})
}
