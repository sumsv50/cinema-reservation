package models

import (
	"time"
)

type Cinema struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;unique"`
	Slug        string    `json:"slug" gorm:"not null;unique"`
	Rows        int       `json:"rows" gorm:"not null"`
	Columns     int       `json:"columns" gorm:"not null"`
	MinDistance int       `json:"min_distance" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCinemaRequest struct {
	Name        string `json:"name" binding:"required,trimmed_min=5"`
	Rows        int    `json:"rows" binding:"required,min=1"`
	Columns     int    `json:"columns" binding:"required,min=1"`
	MinDistance int    `json:"min_distance" binding:"required,min=0"`
}

type Seat struct {
	Row    int `json:"row"`
	Column int `json:"column"`
}

type CinemaLayout struct {
	Cinema Cinema `json:"cinema"`
	Seats  []Seat `json:"seats"`
}
