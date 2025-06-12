package handlers

import (
	"net/http"

	"cinema-reservation/internal/models"
	"cinema-reservation/internal/services"
	"cinema-reservation/internal/utils"

	"github.com/gin-gonic/gin"
)

type CinemaHandler struct {
	cinemaService services.CinemaService
}

func NewCinemaHandler(cinemaService services.CinemaService) *CinemaHandler {
	return &CinemaHandler{cinemaService: cinemaService}
}

func (h *CinemaHandler) CreateLayout(c *gin.Context) {
	var req models.CreateCinemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	cinema, err := h.cinemaService.CreateLayout(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Cinema created successfully", cinema)
}

func (h *CinemaHandler) GetAvailableSeats(c *gin.Context) {
	slug := c.Param("slug")

	layout, err := h.cinemaService.GetAvailableSeats(c.Request.Context(), slug)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Seats retrieved successfully", layout)
}
