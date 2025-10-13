package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secretPtr := flag.String("secret", "", "The secret key used to sign the JWT. Must match the server's JWT_SECRET.")
	expiresInPtr := flag.String("expires-in", "", "The token's validity duration (e.g., \"24h\", \"30m\").")
	flag.Parse()

	if *secretPtr == "" || *expiresInPtr == "" {
		fmt.Println("Error: Both --secret and --expires-in flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	duration, err := time.ParseDuration(*expiresInPtr)
	if err != nil {
		fmt.Printf("Error: Invalid duration format for --expires-in: %v\n", err)
		os.Exit(1)
	}

	// Create the Claims
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		// No Subject or Audience claims are required by the PRD, keeping it minimal.
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the provided secret
	tokenString, err := token.SignedString([]byte(*secretPtr))
	if err != nil {
		fmt.Printf("Error: Failed to sign token: %v\n", err)
		os.Exit(1)
	}

	// Output the token string to stdout as required
	fmt.Print(tokenString)
}
