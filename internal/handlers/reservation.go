package handlers

import (
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

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	reservation, err := h.reservationService.ReserveSeats(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Seats reserved successfully", reservation)
}

func (h *ReservationHandler) CancelSeats(c *gin.Context) {
	var req models.CancelRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	err := h.reservationService.CancelSeats(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Seats canceled successfully", nil)
}
