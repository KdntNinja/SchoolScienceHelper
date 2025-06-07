package utils

import (
	"net/http"
	"os"
	"time"
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
	dur := 24 * time.Hour
	if keepLoggedIn {
		dur = 30 * 24 * time.Hour // 30 days
	}
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    username,
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
	return cookie.Value, nil
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
