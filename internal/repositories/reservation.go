package repositories

import (
	"context"
	"fmt"
	"time"

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

func (r *reservationRepository) LockSeats(ctx context.Context, cinemaID uint, seats []models.ReservedSeat) error {
	for _, seat := range seats {
		key := fmt.Sprintf("seat_lock:%d:%d:%d", cinemaID, seat.Row, seat.Column)
		err := r.redis.Set(ctx, key, "locked", 5*time.Minute).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *reservationRepository) UnlockSeats(ctx context.Context, cinemaID uint, seats []models.ReservedSeat) error {
	for _, seat := range seats {
		key := fmt.Sprintf("seat_lock:%d:%d:%d", cinemaID, seat.Row, seat.Column)
		r.redis.Del(ctx, key)
	}
	return nil
}
