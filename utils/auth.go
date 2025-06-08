package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
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
			log.Error("[getJWKS] AUTH0_DOMAIN not set")
			jwksErr = errors.New("AUTH0_DOMAIN not set")
			return
		}
		jwksURL := "https://" + domain + "/.well-known/jwks.json"
		log.Infof("[getJWKS] Fetching JWKS from %s", jwksURL)
		jwksInstance, jwksErr = keyfunc.Get(jwksURL, keyfunc.Options{
			RefreshInterval:     time.Hour, // refresh every hour
			RefreshErrorHandler: func(err error) { log.Errorf("[getJWKS] JWKS refresh error: %v", err) },
			RefreshTimeout:      10 * time.Second,
			RefreshUnknownKID:   true,
		})
		if jwksErr != nil {
			log.Errorf("[getJWKS] Failed to fetch JWKS: %v", jwksErr)
		}
	})
	if jwksErr != nil {
		log.Error("[getJWKS] Returning error: ", jwksErr)
	}
	return jwksInstance, jwksErr
}

// ValidateAndParseJWT validates the JWT and returns claims if valid
func ValidateAndParseJWT(tokenStr string) (jwt.MapClaims, error) {
	jwks, err := getJWKS()
	if err != nil {
		log.Errorf("[ValidateAndParseJWT] Failed to get JWKS: %v", err)
		return nil, err
	}
	parsed, err := jwt.Parse(tokenStr, jwks.Keyfunc)
	if err != nil {
		log.Warnf("[ValidateAndParseJWT] JWT parse error: %v", err)
		return nil, err
	}
	if !parsed.Valid {
		log.Warn("[ValidateAndParseJWT] Invalid token")
		return nil, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		log.Warn("[ValidateAndParseJWT] Invalid claims type")
		return nil, errors.New("invalid claims")
	}
	// Validate issuer and audience
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	if claims["iss"] != "https://"+domain+"/" {
		log.Warnf("[ValidateAndParseJWT] Invalid issuer: %v", claims["iss"])
		return nil, errors.New("invalid issuer")
	}
	aud, ok := claims["aud"]
	if !ok {
		log.Warn("[ValidateAndParseJWT] Missing audience claim")
		return nil, errors.New("missing audience")
	}
	switch v := aud.(type) {
	case string:
		if v != clientID {
			log.Warnf("[ValidateAndParseJWT] Invalid audience: %v", v)
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
			log.Warnf("[ValidateAndParseJWT] Audience does not include clientID: %v", v)
			return nil, errors.New("invalid audience")
		}
	default:
		log.Warnf("[ValidateAndParseJWT] Invalid audience type: %T", v)
		return nil, errors.New("invalid audience type")
	}
	return claims, nil
}

// GetJWTFromRequest extracts the JWT from the request, checking both the Authorization header and cookies.
func GetJWTFromRequest(r *http.Request) string {
	if cookie, err := r.Cookie("auth_token"); err == nil {
		return cookie.Value
	} else if err != http.ErrNoCookie {
		log.Warnf("[GetJWTFromRequest] Error reading cookie: %v", err)
	}
	header := r.Header.Get("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}

// GetUserIDFromRequest extracts the user ID (sub claim) from a validated Auth0 JWT in the request.
// This function performs full signature and claims validation using Auth0 JWKS.
func GetUserIDFromRequest(r *http.Request) (string, error) {
	tokenStr := GetJWTFromRequest(r)
	log.Infof("[GetUserIDFromRequest] Extracting user from JWT (token present: %v) from %s %s", tokenStr != "", r.Method, r.URL.Path)
	if tokenStr == "" {
		log.Warnf("[GetUserIDFromRequest] Missing token from %s %s", r.Method, r.URL.Path)
		return "", fmt.Errorf("missing token")
	}
	claims, err := ValidateAndParseJWT(tokenStr)
	if err != nil {
		log.Warnf("[GetUserIDFromRequest] Invalid JWT: %v", err)
		return "", err
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		log.Warnf("[GetUserIDFromRequest] Missing sub claim in JWT")
		return "", fmt.Errorf("missing sub claim")
	}
	log.Infof("[GetUserIDFromRequest] User ID extracted: %s", userID)
	return userID, nil
}
