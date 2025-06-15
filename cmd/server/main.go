package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"KdnSite/assets"
	"KdnSite/internal/handlers"
	"KdnSite/internal/projects"
	"KdnSite/internal/quizzes"
	errorpages "KdnSite/ui/pages/error"
	legalpages "KdnSite/ui/pages/legal"
	publicpages "KdnSite/ui/pages/public"
	userpages "KdnSite/ui/pages/user"

	log "github.com/sirupsen/logrus"
)

func main() {
	setupLogging()
	checkEnvVars()
	db := setupDatabase()
	defer db.Close()

	mux := http.NewServeMux()
	registerStaticRoutes(mux)
	registerLegalRoutes(mux)
	registerAuthRoutes(mux)
	registerUserRoutes(mux)
	SetupAssetsRoutes(mux)
	registerAPIRoutes(mux)

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

func setupLogging() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)
}

func checkEnvVars() {
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	if domain == "" {
		log.Fatal("AUTH0_DOMAIN environment variable is not set!")
	}
	if clientID == "" {
		log.Fatal("AUTH0_CLIENT_ID environment variable is not set!")
	}
	dbURL := os.Getenv("NEON_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("NEON_DATABASE_URL environment variable is not set!")
	}
}

func setupDatabase() *sql.DB {
	dbURL := os.Getenv("NEON_DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to NeonDB: %v", err)
	}
	return db
}

func registerStaticRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			err := errorpages.NotFound().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (NotFound): %v", err)
			}
			return
		}
		domain := os.Getenv("AUTH0_DOMAIN")
		clientID := os.Getenv("AUTH0_CLIENT_ID")
		err := publicpages.Landing(domain, clientID).Render(r.Context(), w)
		if err != nil {
			log.Errorf("Render error (Landing): %v", err)
		}
	})
}

func registerLegalRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		err := legalpages.Terms().Render(r.Context(), w)
		if err != nil {
			log.Errorf("Render error (Terms): %v", err)
		}
	})
	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		err := legalpages.Privacy().Render(r.Context(), w)
		if err != nil {
			log.Errorf("Render error (Privacy): %v", err)
		}
	})
}

func registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/check", handlers.AuthStatusHandler)
	mux.HandleFunc("/api/auth/callback", handlers.HandleAuthCallback)
}

func registerUserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/dash", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages.Dash().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (Dash): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/forbidden", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		err := errorpages.Forbidden().Render(r.Context(), w)
		if err != nil {
			log.Errorf("Render error (Forbidden): %v", err)
		}
	})
	mux.HandleFunc("/badrequest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		err := errorpages.BadRequest().Render(r.Context(), w)
		if err != nil {
			log.Errorf("Render error (BadRequest): %v", err)
		}
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		err := errorpages.InternalServerError().Render(r.Context(), w)
		if err != nil {
			log.Errorf("Render error (InternalServerError): %v", err)
		}
	})
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

func registerAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projects.ListProjects(w, r)
		case http.MethodPost:
			projects.CreateProject(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/quizzes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			quizzes.ListQuizzes(w, r)
		case http.MethodPost:
			quizzes.CreateQuiz(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
