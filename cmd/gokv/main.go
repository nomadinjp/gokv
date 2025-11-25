package main

import (
	"fmt"
	"log"
	"os"

	"gokv/internal/handler"
	"gokv/internal/storage"

	"github.com/gin-gonic/gin"
)

// setupRouter initializes the Gin engine and sets up the routes.
func setupRouter(store *storage.Storage) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery()) // Use default Gin middleware

	// Health check endpoint (optional but good practice)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Protected API routes
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware())
	{
		kvHandler := handler.NewHandler(store)
		api.POST("/:bucket/:key", kvHandler.PostHandler)
		api.GET("/:bucket/:key", kvHandler.GetHandler)
		api.DELETE("/:bucket/:key", kvHandler.DeleteHandler)
		api.GET("/_list", kvHandler.ListHandler)
	}

	return r
}

func main() {
	// --- Configuration and Initialization ---
	// Read PORT and GIN_MODE environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// Initialize Storage (handles DB_PATH env var)
	store, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close() // Ensure database is closed on exit

	// --- Router Setup ---
	r := setupRouter(store)

	// --- Start Server ---
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s (Mode: %s)", addr, gin.Mode())
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
