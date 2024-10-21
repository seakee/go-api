// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package jwt provides functionality for generating and parsing JSON Web Tokens (JWT)
// specifically for application authentication in the go-api project.
package jwt

import (
	"github.com/seakee/go-api/app/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/seakee/go-api/app/model/auth"
)

// ServerClaims represents the custom claims structure for the JWT.
// It extends the standard RegisteredClaims with application-specific fields.
type ServerClaims struct {
	ID      uint   `json:"id"`       // Unique identifier for the app
	AppName string `json:"app_name"` // Name of the application
	AppID   string `json:"app_id"`   // Unique ID of the application
	jwt.RegisteredClaims
}

// GenerateAppToken creates a new JWT token for an application.
//
// Parameters:
//   - App: A pointer to the auth.App struct containing application details.
//   - expireTime: The duration for which the token should be valid, in seconds.
//
// Returns:
//   - token: The generated JWT token as a string.
//   - err: An error if token generation fails, nil otherwise.
//
// Example:
//
//	app := &auth.App{ID: 1, AppName: "MyApp", AppID: "APP123"}
//	token, err := GenerateAppToken(app, 3600) // Token valid for 1 hour
//	if err != nil {
//	    log.Fatalf("Failed to generate token: %v", err)
//	}
func GenerateAppToken(App *auth.App, expireTime time.Duration) (token string, err error) {
	// Calculate the expiration time
	expTime := time.Now().Add(expireTime * time.Second)

	// Create the claims
	claims := ServerClaims{
		ID:      App.ID,
		AppName: App.AppName,
		AppID:   App.AppID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-api",
		},
	}

	// Create a new token object, specifying signing method and the claims
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	jwtSecret := []byte(config.Get().System.JwtSecret)
	return tokenClaims.SignedString(jwtSecret)
}

// ParseAppAuth parses and validates a JWT token string.
//
// Parameters:
//   - token: The JWT token string to be parsed and validated.
//
// Returns:
//   - *ServerClaims: A pointer to the parsed ServerClaims if the token is valid.
//   - error: An error if parsing fails or the token is invalid, nil otherwise.
//
// Example:
//
//	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//	claims, err := ParseAppAuth(tokenString)
//	if err != nil {
//	    log.Fatalf("Failed to parse token: %v", err)
//	}
//	fmt.Printf("App ID: %s\n", claims.AppID)
func ParseAppAuth(token string) (*ServerClaims, error) {
	jwtSecret := []byte(config.Get().System.JwtSecret)

	// Parse the token
	tokenClaims, err := jwt.ParseWithClaims(token, &ServerClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	// Check if the token is valid
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*ServerClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}
