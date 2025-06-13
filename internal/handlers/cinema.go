package handlers

import (
	"net/http"
	"strconv"

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
	nStr := c.Query("number_of_seats")
	numberOfSeats, err := strconv.Atoi(nStr)
	if err != nil || numberOfSeats <= 0 {
		numberOfSeats = 1
	}

	available, err := h.cinemaService.GetAvailableSeats(c.Request.Context(), slug, numberOfSeats)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Available seats retrieved successfully", available)
}

func (h *CinemaHandler) CheckAvailableSeats(c *gin.Context) {
	slug := c.Param("slug")
	var req models.CheckSeatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	available, err := h.cinemaService.CheckAvailableSeats(c.Request.Context(), slug, &req)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Check seats successfully", available)
}
