package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Code    string      `json:"code,omitempty"`
}

type ErrorMapping struct {
	StatusCode int
	Message    string
	Code       string
}

var errorMappings = map[error]ErrorMapping{
	// Cinema errors
	ErrCinemaNotFound:      {http.StatusNotFound, "Cinema not found", "CINEMA_NOT_FOUND"},
	ErrCinemaAlreadyExists: {http.StatusConflict, "Cinema with this name already exists", "CINEMA_EXISTS"},

	// Reservation errors
	ErrSeatsAlreadyReserved: {http.StatusConflict, "One or more seats are already reserved", "SEATS_RESERVED"},
	ErrSeatsNotAvailable:    {http.StatusConflict, "Selected seats are not available", "SEATS_NOT_AVAILABLE"},
	ErrInvalidSeatPosition:  {http.StatusBadRequest, "Invalid seat position", "INVALID_SEAT_POSITION"},
	ErrMinDistanceViolation: {http.StatusBadRequest, "Seat selection violates minimum distance requirement", "MIN_DISTANCE_VIOLATION"},

	// General errors
	ErrInvalidInput:       {http.StatusBadRequest, "Invalid input provided", "INVALID_INPUT"},
	ErrInternalServer:     {http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR"},
	ErrDatabaseConnection: {http.StatusServiceUnavailable, "Service temporarily unavailable", "SERVICE_UNAVAILABLE"},
	ErrRateLimitExceeded:  {http.StatusTooManyRequests, "Rate limit exceeded", "RATE_LIMIT_EXCEEDED"},
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, err error) {
	response := Response{
		Success: false,
	}
	var statusCode int

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		response.Message = "Validation failed"
		response.Code = "VALIDATION_ERROR"
		statusCode = http.StatusBadRequest

		var details []string
		for _, fieldErr := range validationErrs {
			details = append(details, fmt.Sprintf("Field '%s' failed validation: %s", fieldErr.Field(), fieldErr.Tag()))
		}
		response.Data = map[string]interface{}{"validation_errors": details}

		c.JSON(statusCode, response)
		return
	}

	mapping, exists := errorMappings[err]
	if exists {
		response.Message = mapping.Message
		response.Code = mapping.Code
		statusCode = mapping.StatusCode
	} else {
		response.Message = "An unexpected error occurred"
		response.Code = "UNKNOWN_ERROR"
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, response)
}
