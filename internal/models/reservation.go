package models

import (
	"fmt"
	"strings"
	"time"
)

type Reservation struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	CinemaID   uint           `json:"cinema_id" gorm:"not null"`
	Note       string         `json:"note"`
	ReservedAt time.Time      `json:"reserved_at"`
	Cinema     Cinema         `json:"-" gorm:"foreignKey:CinemaID"`
	Seats      []ReservedSeat `json:"seats,omitempty" gorm:"foreignKey:ReservationID;constraint:OnDelete:CASCADE"`
}

type ReservedSeat struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	CinemaID      uint   `json:"cinema_id" gorm:"not null;uniqueIndex:idx_cinema_seat"`
	ReservationID uint   `json:"reservation_id" gorm:"not null"`
	Row           int    `json:"row" gorm:"not null;uniqueIndex:idx_cinema_seat"`
	Column        int    `json:"column" gorm:"not null;uniqueIndex:idx_cinema_seat"`
	Cinema        Cinema `json:"-" gorm:"foreignKey:CinemaID"`
}

type SeatRequest struct {
	Row    int `json:"row" binding:"required,min=0"`
	Column int `json:"column" binding:"required,min=0"`
}

type ReservationRequest struct {
	CinemaSlug string        `json:"cinema_slug" binding:"required"`
	Note       string        `json:"note"`
	Seats      []SeatRequest `json:"seats" binding:"required,dive,required"`
}

type ReservedSeats []ReservedSeat

func (seats ReservedSeats) String() string {
	var seatStrings []string
	for _, seat := range seats {
		seatStrings = append(seatStrings, fmt.Sprintf("R%dC%d", seat.Row, seat.Column))
	}

	return strings.Join(seatStrings, ", ")
}
