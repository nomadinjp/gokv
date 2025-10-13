package handler

import (
	"fmt"
	"io"
	"net/http"

	"gokv/internal/storage"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

// Handler holds the storage instance and provides HTTP handler methods.
type Handler struct {
	Store *storage.Storage
}

// NewHandler creates a new Handler instance.
func NewHandler(store *storage.Storage) *Handler {
	return &Handler{Store: store}
}

// PostHandler handles POST /:bucket/:key requests (Create or Update).
func (h *Handler) PostHandler(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bucket and key must not be empty"})
		return
	}

	// Read the entire raw request body as the value
	value, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	if err := h.Store.Set(bucket, key, value); err != nil {
		// Log the error internally
		fmt.Printf("Error setting key %s:%s: %v\n", bucket, key, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while storing data"})
		return
	}

	// Success response
	c.Status(http.StatusOK)
}

// GetHandler handles GET /:bucket/:key requests (Read).
func (h *Handler) GetHandler(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bucket and key must not be empty"})
		return
	}

	value, err := h.Store.Get(bucket, key)

	if err != nil {
		if err == badger.ErrKeyNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Key not found"})
			return
		}
		// Log the error internally
		fmt.Printf("Error getting key %s:%s: %v\n", bucket, key, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while retrieving data"})
		return
	}

	// Success response: 200 OK, raw data as body, Content-Type application/octet-stream
	c.Data(http.StatusOK, "application/octet-stream", value)
}
