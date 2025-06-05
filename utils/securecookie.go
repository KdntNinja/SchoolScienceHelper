package utils

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/securecookie"
)

var (
	cookieHashKey  = []byte(getEnvOrDefault("SESSION_HASH_KEY", "very-secret-key-123"))
	cookieBlockKey = []byte(getEnvOrDefault("SESSION_BLOCK_KEY", "a-lot-of-random-bytes"))
	sc             = securecookie.New(cookieHashKey, cookieBlockKey)
)

func getEnvOrDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

const sessionCookieName = "session"

func SetSession(w http.ResponseWriter, username string, keepLoggedIn bool) error {
	encoded, err := sc.Encode(sessionCookieName, map[string]string{"username": username})
	if err != nil {
		return err
	}
	dur := 24 * time.Hour
	if keepLoggedIn {
		dur = 30 * 24 * time.Hour // 30 days
	}
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(dur),
	}
	// For production, set Secure: true
	w.Header().Add("Set-Cookie", cookie.String())
	return nil
}

func GetSessionUsername(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	value := make(map[string]string)
	if err := sc.Decode(sessionCookieName, cookie.Value, &value); err != nil {
		return "", err
	}
	return value["username"], nil
}

func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	w.Header().Add("Set-Cookie", cookie.String())
}
