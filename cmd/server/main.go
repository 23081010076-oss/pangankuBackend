// Penjelasan file:
// Lokasi: cmd/server/main.go
// Bagian: entrypoint
// File: main
// Fungsi utama: File ini menjadi titik masuk server atau utilitas backend.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/panganku/backend/internal/config"
	"github.com/panganku/backend/internal/handlers"
	"github.com/panganku/backend/internal/middleware"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file tidak ditemukan, menggunakan environment variables")
	}

	// Connect to database and Redis
	db := config.ConnectDB()
	rdb := config.ConnectRedis()

	// Run migrations
	config.AutoMigrate(db)

	// Seed initial data
	config.SeedData(db)
	if os.Getenv("APP_ENV") != "production" || os.Getenv("SEED_DUMMY_DATA") == "true" {
		config.SeedDummyData(db)
	}

	// Setup Gin router
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Apply CORS middleware (must be first for preflight requests)
	router.Use(middleware.CORS())

	// Apply security headers middleware
	router.Use(middleware.SecurityHeaders())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now()})
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, rdb)
	hargaHandler := handlers.NewHargaHandler(db, rdb)
	uploadHandler := handlers.NewUploadHandler()
	komoditasHandler := handlers.NewKomoditasHandler(db, rdb)
	kecamatanHandler := handlers.NewKecamatanHandler(db)
	telegramHandler := handlers.NewTelegramHandler(db, rdb)
	stokHandler := handlers.NewStokHandler(db)
	laporanHandler := handlers.NewLaporanHandler(db, rdb, telegramHandler)
	distribusiHandler := handlers.NewDistribusiHandler(db)
	notifikasiHandler := handlers.NewNotifikasiHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(db)
	userHandler := handlers.NewUserHandler(db)
	luasLahanHandler := handlers.NewLuasLahanHandler(db)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public webhook routes
		telegramHandler.RegisterRoutes(v1)

		// Public routes (no auth required)
		authHandler.RegisterRoutes(v1)
		hargaHandler.RegisterRoutes(v1)
		komoditasHandler.RegisterRoutes(v1)
		kecamatanHandler.RegisterRoutes(v1)
		stokHandler.RegisterRoutes(v1)
		luasLahanHandler.RegisterRoutes(v1)
		analyticsHandler.RegisterRoutes(v1)

		// Authenticated routes
		authRequired := v1.Group("", middleware.JWTAuth(rdb))
		{
			uploadHandler.RegisterRoutes(authRequired)
			laporanHandler.RegisterRoutes(authRequired)
			distribusiHandler.RegisterRoutes(authRequired)
			notifikasiHandler.RegisterRoutes(authRequired)
			userHandler.RegisterRoutes(authRequired)

			// Admin / petugas only
			adminOnly := authRequired.Group("", middleware.RequireRole("admin", "petugas"))
			{
				adminOnly.POST("/komoditas", komoditasHandler.CreateKomoditas)
				adminOnly.PUT("/komoditas/:id", komoditasHandler.UpdateKomoditas)
				adminOnly.DELETE("/komoditas/:id", komoditasHandler.DeleteKomoditas)

				adminOnly.POST("/kecamatan", kecamatanHandler.CreateKecamatan)
				adminOnly.PUT("/kecamatan/:id", kecamatanHandler.UpdateKecamatan)
				adminOnly.DELETE("/kecamatan/:id", kecamatanHandler.DeleteKecamatan)

				adminOnly.POST("/stok", stokHandler.CreateOrUpdateStok)
				adminOnly.DELETE("/stok/:id", stokHandler.DeleteStok)

				adminOnly.POST("/luas-lahan", luasLahanHandler.CreateOrUpdateLuasLahan)
				adminOnly.DELETE("/luas-lahan/:id", luasLahanHandler.DeleteLuasLahan)

				adminOnly.POST("/distribusi", distribusiHandler.CreateDistribusi)
				adminOnly.PUT("/distribusi/:id/status", distribusiHandler.UpdateDistribusiStatus)
				adminOnly.DELETE("/distribusi/:id", distribusiHandler.DeleteDistribusi)
			}

			// Admin only
			adminRoute := authRequired.Group("", middleware.RequireRole("admin"))
			{
				userHandler.RegisterAdminRoutes(adminRoute)
			}
		}
	}

	// Serve static files (uploads)
	router.Static("/uploads", "./uploads")

	// Get port from env. Render/Koyeb/Fly biasanya menyediakan PORT,
	// sedangkan local/docker project ini memakai APP_PORT.
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	// Setup HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("ðŸš€ Server berjalan di http://localhost:%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	// Close database connection
	sqlDB, _ := db.DB()
	sqlDB.Close()

	// Close Redis connection
	rdb.Close()

	log.Println("Server stopped gracefully")
}
