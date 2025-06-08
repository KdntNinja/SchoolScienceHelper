package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

var (
	jwksInstance *keyfunc.JWKS
	jwksOnce     sync.Once
	jwksErr      error
)

// getJWKS fetches and caches the JWKS from Auth0
func getJWKS() (*keyfunc.JWKS, error) {
	jwksOnce.Do(func() {
		domain := os.Getenv("AUTH0_DOMAIN")
		if domain == "" {
			jwksErr = errors.New("AUTH0_DOMAIN not set")
			return
		}
		jwksURL := "https://" + domain + "/.well-known/jwks.json"
		jwksInstance, jwksErr = keyfunc.Get(jwksURL, keyfunc.Options{
			RefreshInterval:     time.Hour,          // refresh every hour
			RefreshErrorHandler: func(err error) {}, // Log error
			RefreshTimeout:      10 * time.Second,
			RefreshUnknownKID:   true,
		})
	})
	return jwksInstance, jwksErr
}

// ValidateAndParseJWT validates the JWT and returns claims if valid
func ValidateAndParseJWT(tokenStr string) (jwt.MapClaims, error) {
	jwks, err := getJWKS()
	if err != nil {
		return nil, err
	}
	parsed, err := jwt.Parse(tokenStr, jwks.Keyfunc)
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	// Validate issuer and audience
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	if claims["iss"] != "https://"+domain+"/" {
		return nil, errors.New("invalid issuer")
	}
	aud, ok := claims["aud"]
	if !ok {
		return nil, errors.New("missing audience")
	}
	switch v := aud.(type) {
	case string:
		if v != clientID {
			return nil, errors.New("invalid audience")
		}
	case []interface{}:
		found := false
		for _, a := range v {
			if s, ok := a.(string); ok && s == clientID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid audience")
		}
	default:
		return nil, errors.New("invalid audience type")
	}
	return claims, nil
}

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

// Update GetUserIDFromRequest to use validation
func GetUserIDFromRequest(r *http.Request) (string, error) {
	tokenStr := GetJWTFromRequest(r)
	if tokenStr == "" {
		return "", fmt.Errorf("missing token")
	}
	claims, err := ValidateAndParseJWT(tokenStr)
	if err != nil {
		return "", err
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing sub claim")
	}
	return userID, nil
}
