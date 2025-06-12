package repositories

import (
	"context"

	"cinema-reservation/internal/models"

	"gorm.io/gorm"
)

type cinemaRepository struct {
	db *gorm.DB
}

func NewCinemaRepository(db *gorm.DB) CinemaRepository {
	return &cinemaRepository{db: db}
}

func (r *cinemaRepository) Create(ctx context.Context, cinema *models.Cinema) error {
	return r.db.WithContext(ctx).Create(cinema).Error
}

func (r *cinemaRepository) GetBySlug(ctx context.Context, slug string) (*models.Cinema, error) {
	var cinema models.Cinema
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&cinema).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cinema, nil
}

func (r *cinemaRepository) GetReservedSeats(ctx context.Context, cinemaID uint) ([]models.ReservedSeat, error) {
	var seats []models.ReservedSeat
	err := r.db.WithContext(ctx).Where("cinema_id = ?", cinemaID).Find(&seats).Error
	return seats, err
}

func (r *cinemaRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Cinema{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
