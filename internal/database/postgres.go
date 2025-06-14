package database

import (
	"cinema-reservation/internal/models"
	"time"

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

	// Get the underlying sql.DB instance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool settings
	sqlDB.SetMaxOpenConns(80)                 // Maximum number of open connections
	sqlDB.SetMaxIdleConns(20)                 // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime
	sqlDB.SetConnMaxIdleTime(5 * time.Minute) // Maximum idle time

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
