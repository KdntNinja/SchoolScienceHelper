package main

import (
	"fmt"
	"net/http"
	"os"

	"SchoolScienceHelper/assets"
	"SchoolScienceHelper/internal/handlers"
	errorpages "SchoolScienceHelper/ui/pages/error"
	publicpages "SchoolScienceHelper/ui/pages/public"
	userpages "SchoolScienceHelper/ui/pages/user"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)

	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	if domain == "" {
		log.Warn("AUTH0_DOMAIN environment variable is not set!")
	}
	if clientID == "" {
		log.Warn("AUTH0_CLIENT_ID environment variable is not set!")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"AUTH0_DOMAIN":%q,"AUTH0_CLIENT_ID":%q}`,
			domain, clientID)
	})
	SetupAssetsRoutes(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			errorpages.NotFound().Render(r.Context(), w)
			return
		}
		publicpages.Landing().Render(r.Context(), w)
	})
	mux.HandleFunc("/api/auth/callback", handlers.HandleAuthCallback)
	mux.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintln(w, "Terms page not implemented")
	})
	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintln(w, "Privacy page not implemented")
	})
	mux.HandleFunc("/dash", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userpages.Dash().Render(r.Context(), w)
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/forbidden", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		errorpages.Forbidden().Render(r.Context(), w)
	})
	mux.HandleFunc("/badrequest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorpages.BadRequest().Render(r.Context(), w)
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		errorpages.InternalServerError().Render(r.Context(), w)
	})

	hstsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if os.Getenv("GO_ENV") == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			}
			next.ServeHTTP(w, r)
		})
	}

	handler := hstsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	log.Infof("Server running on :%s", port)

	if os.Getenv("GO_ENV") == "production" {
		go func() {
			http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
			}))
		}()
	}

	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		return
	}
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
