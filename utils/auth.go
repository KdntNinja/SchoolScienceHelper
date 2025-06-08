package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// GetUserIDFromAuthHeader extracts the user ID (sub claim) from the Authorization header.
func GetUserIDFromAuthHeader(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", fmt.Errorf("missing bearer token")
	}
	tokenStr := strings.TrimPrefix(header, "Bearer ")
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing sub claim")
	}
	return userID, nil
}

// Helper: get JWT from cookie or Authorization header
func GetJWTFromRequest(r *http.Request) string {
	// Try cookie first
	if cookie, err := r.Cookie("auth_token"); err == nil {
		return cookie.Value
	}
	header := r.Header.Get("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}

// Update GetUserIDFromAuthHeader to use GetJWTFromRequest
func GetUserIDFromRequest(r *http.Request) (string, error) {
	tokenStr := GetJWTFromRequest(r)
	if tokenStr == "" {
		return "", fmt.Errorf("missing token")
	}
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing sub claim")
	}
	return userID, nil
}
