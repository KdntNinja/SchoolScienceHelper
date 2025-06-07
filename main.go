package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/KdntNinja/ScratchClone/assets"
	"github.com/KdntNinja/ScratchClone/ui/pages"
	"github.com/a-h/templ"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Logging setup (expandable)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)

	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	if domain == "" {
		log.Warn("AUTH0_DOMAIN environment variable is not set!")
	} else {
		log.Infof("AUTH0_DOMAIN: %s", domain)
	}
	if clientID == "" {
		log.Warn("AUTH0_CLIENT_ID environment variable is not set!")
	} else {
		log.Infof("AUTH0_CLIENT_ID: %s", clientID)
	}

	mux := http.NewServeMux()
	SetupAssetsRoutes(mux)

	// Auth0-protected route middleware (pseudo, update as needed for your Auth0 integration)
	requireAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Example: check for Auth0 ID token in cookie or Authorization header
			// If not present/valid, redirect to /auth
			idToken := r.Header.Get("Authorization")
			if idToken == "" {
				http.Redirect(w, r, "/auth", http.StatusSeeOther)
				return
			}
			// Optionally: validate the token here (implementation depends on your Auth0 setup)
			next.ServeHTTP(w, r)
		})
	}

	// Public routes
	mux.Handle("GET /", templ.Handler(pages.Landing()))
	mux.Handle("GET /auth", templ.Handler(pages.Auth()))
	mux.Handle("GET /terms", templ.Handler(pages.Terms()))
	mux.Handle("GET /privacy", templ.Handler(pages.Privacy()))

	// Protected routes
	mux.Handle("GET /dash", requireAuth(templ.Handler(pages.Dash())))
	// Add more protected routes as needed, using requireAuth

	// Serve Auth0 config for frontend
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"AUTH0_DOMAIN":%q,"AUTH0_CLIENT_ID":%q}`, domain, clientID)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	fmt.Printf("Server running on :%s\n", port)
	http.ListenAndServe(":"+port, mux)
}

func SetupAssetsRoutes(mux *http.ServeMux) {
	var isDevelopment = os.Getenv("GO_ENV") != "production"

	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isDevelopment {
			w.Header().Set("Cache-Control", "no-store")
		}

		var fs http.Handler
		if isDevelopment {
			fs = http.FileServer(http.Dir("./assets"))
		} else {
			fs = http.FileServer(http.FS(assets.Assets))
		}

		fs.ServeHTTP(w, r)
	})

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))
}
