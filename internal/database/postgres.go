package database

import (
	"cinema-reservation/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate
	err = db.AutoMigrate(
		&models.Cinema{},
		&models.Reservation{},
		&models.ReservedSeat{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
