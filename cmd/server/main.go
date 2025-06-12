package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"cinema-reservation/internal/config"
	"cinema-reservation/internal/database"
	"cinema-reservation/internal/handlers"
	"cinema-reservation/internal/middleware"
	"cinema-reservation/internal/repositories"
	"cinema-reservation/internal/services"
	validators "cinema-reservation/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize databases
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	redis, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validators.RegisterCustomValidators(v)
		log.Println("Custom validators registered successfully")
	} else {
		log.Println("Failed to register custom validators")
	}

	// Initialize repositories
	cinemaRepo := repositories.NewCinemaRepository(db)
	reservationRepo := repositories.NewReservationRepository(db, redis)

	// Initialize services
	cinemaService := services.NewCinemaService(cinemaRepo)
	reservationService := services.NewReservationService(reservationRepo, cinemaRepo, redis)

	// Initialize handlers
	cinemaHandler := handlers.NewCinemaHandler(cinemaService)
	reservationHandler := handlers.NewReservationHandler(reservationService)
	healthHandler := handlers.NewHealthHandler(db, redis)

	// Setup router
	router := setupRouter(cinemaHandler, reservationHandler, healthHandler, redis)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func setupRouter(
	cinemaHandler *handlers.CinemaHandler,
	reservationHandler *handlers.ReservationHandler,
	healthHandler *handlers.HealthHandler,
	redis *redis.Client,
) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Rate limiting middleware (100 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(redis, 100, time.Minute)
	router.Use(rateLimiter.Middleware())

	// Health check (no rate limiting)
	router.GET("/health", healthHandler.Check)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Cinema routes
		cinemas := v1.Group("/cinemas")
		{
			cinemas.POST("", cinemaHandler.CreateLayout)
			cinemas.GET("/:slug/seats", cinemaHandler.GetAvailableSeats)
		}

		// Reservation routes
		reservations := v1.Group("/reservations")
		{
			reservations.POST("", reservationHandler.ReserveSeats)
		}
	}

	return router
}
