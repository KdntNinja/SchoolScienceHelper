package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/KdntNinja/ScratchClone/assets"
	errorpages "github.com/KdntNinja/ScratchClone/ui/pages/error"
	publicpages "github.com/KdntNinja/ScratchClone/ui/pages/public"
	userpages "github.com/KdntNinja/ScratchClone/ui/pages/user"
	"github.com/KdntNinja/ScratchClone/utils"
	log "github.com/sirupsen/logrus"
)

// Auth middleware for all user related routes
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := utils.GetUserIDFromRequest(r)
		if err != nil || userID == "" {
			log.Warnf("[requireAuth] Unauthorized access: %v, remote=%s, path=%s, cookie=%v", err, r.RemoteAddr, r.URL.Path, r.Header.Get("Cookie"))
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}
		log.Infof("[requireAuth] Authenticated user: %s, remote=%s, path=%s", userID, r.RemoteAddr, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// =====================
	// Logging & Environment
	// =====================
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

	// =============
	// HTTP Handlers
	// =============
	mux := http.NewServeMux()

	// --- Public Config & Assets ---
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"AUTH0_DOMAIN":%q,"AUTH0_CLIENT_ID":%q}`,
			domain, clientID)
	})
	SetupAssetsRoutes(mux)

	// --- Public Pages ---
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
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		publicpages.Auth().Render(r.Context(), w)
	})
	mux.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintln(w, "Terms page not implemented")
	})
	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintln(w, "Privacy page not implemented")
	})
	// --- User Pages (Require Auth) ---
	mux.HandleFunc("/dash", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userpages.Dash().Render(r.Context(), w)
		})).ServeHTTP(w, r)
	})

	// --- API: User (Require Auth) ---
	mux.Handle("/api/user/profile", requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			utils.HandleUserProfile(w, r)
		} else if r.Method == http.MethodPost {
			utils.HandleUserProfileUpdate(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/user/delete", requireAuth(http.HandlerFunc(utils.HandleUserDelete)))

	// --- API: Auth Callback (Public) ---
	mux.HandleFunc("/api/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		utils.HandleAuthCallback(w, r)
	})

	// --- Error Pages ---
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

	// =============
	// DB Setup
	// =============
	db := utils.SetupDB()
	utils.SetDB(db)

	// =============
	// Middleware & Server
	// =============
	// HSTS middleware
	hstsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only set HSTS in production
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

	// HTTP to HTTPS redirect in production
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
