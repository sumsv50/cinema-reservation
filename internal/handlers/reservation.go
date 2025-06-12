package handlers

import (
	"fmt"
	"net/http"

	"cinema-reservation/internal/models"
	"cinema-reservation/internal/services"
	"cinema-reservation/internal/utils"

	"github.com/gin-gonic/gin"
)

type ReservationHandler struct {
	reservationService services.ReservationService
}

func NewReservationHandler(reservationService services.ReservationService) *ReservationHandler {
	return &ReservationHandler{reservationService: reservationService}
}

func (h *ReservationHandler) ReserveSeats(c *gin.Context) {
	var req models.ReservationRequest

	// Debug: Print the raw request body
	fmt.Printf("Content-Type: %s\n", c.GetHeader("Content-Type"))
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	// Debug: Print the parsed request
	fmt.Printf("Parsed request: %+v\n", req)

	reservation, err := h.reservationService.ReserveSeats(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Seats reserved successfully", reservation)
}
