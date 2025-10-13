package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware returns a Gin middleware function that validates JWT tokens.
func AuthMiddleware() gin.HandlerFunc {
	// Read the secret once when the middleware is initialized
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// As per PRD, JWT_SECRET is required. Panic if not set.
		panic("JWT_SECRET environment variable is not set. Cannot start server.")
	}
	secretKey := []byte(secret)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format. Must be 'Bearer <token>'"})
			return
		}

		tokenString := parts[1]

		// Key function to provide the secret for validation
		keyFunc := func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		}

		// Parse and validate the token with a small leeway for clock skew
		token, err := jwt.Parse(tokenString, keyFunc, jwt.WithLeeway(5*time.Second))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid or expired token: %v", err)})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is invalid"})
			return
		}

		// Token is valid, proceed to the next handler
		c.Next()
	}
}
