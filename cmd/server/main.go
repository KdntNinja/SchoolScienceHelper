package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"SchoolScienceHelper/assets"
	"SchoolScienceHelper/internal/aqa"
	"SchoolScienceHelper/internal/handlers"
	errorpages "SchoolScienceHelper/ui/pages/error"
	legalpages "SchoolScienceHelper/ui/pages/legal"
	publicpages "SchoolScienceHelper/ui/pages/public"
	userpages "SchoolScienceHelper/ui/pages/user"
	science "SchoolScienceHelper/ui/pages/user/science"

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

	dbURL := os.Getenv("NEON_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("NEON_DATABASE_URL environment variable is not set!")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to NeonDB: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

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
		domain := os.Getenv("AUTH0_DOMAIN")
		clientID := os.Getenv("AUTH0_CLIENT_ID")
		publicpages.Landing(domain, clientID).Render(r.Context(), w)
	})
	mux.HandleFunc("/api/auth/check", handlers.AuthStatusHandler)
	mux.HandleFunc("/api/auth/callback", handlers.HandleAuthCallback)
	mux.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		legalpages.Terms().Render(r.Context(), w)
	})
	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		legalpages.Privacy().Render(r.Context(), w)
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

	// Register multi-board/tier endpoints
	// Supported boards: aqa, ocr, edexcel
	// Supported tiers: foundation, higher, separated_foundation, separated_higher
	boards := []string{"aqa", "ocr", "edexcel"}
	tiers := []string{"foundation", "higher", "separated_foundation", "separated_higher"}
	for _, board := range boards {
		for _, tier := range tiers {
			mux.HandleFunc("/api/"+board+"/"+tier+"/spec", aqa.SpecsAPI(db))
			mux.HandleFunc("/api/"+board+"/"+tier+"/papers", aqa.PapersAPI(db))
			mux.HandleFunc("/api/"+board+"/"+tier+"/questions", aqa.QuestionsAPI(db))
			mux.HandleFunc("/api/"+board+"/"+tier+"/revision", aqa.RevisionAPI(db))
		}
	}

	mux.HandleFunc("/user/science/spec", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("[ScienceSpecPage] user=%v, path=%s", r.Context().Value("user"), r.URL.Path)
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("Rendering ScienceSpecPage for user=%v", r.Context().Value("user"))
			err := science.ScienceSpecPage().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ScienceSpecPage): %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/science/papers", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("[SciencePapersPage] user=%v, path=%s", r.Context().Value("user"), r.URL.Path)
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("Rendering SciencePapersPage for user=%v", r.Context().Value("user"))
			err := science.SciencePapersPage().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (SciencePapersPage): %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/science/questions", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("[ScienceQuestionsPage] user=%v, path=%s", r.Context().Value("user"), r.URL.Path)
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("Rendering ScienceQuestionsPage for user=%v", r.Context().Value("user"))
			err := science.ScienceQuestionsPage().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ScienceQuestionsPage): %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/science/revision", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("[ScienceRevisionPage] user=%v, path=%s", r.Context().Value("user"), r.URL.Path)
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("Rendering ScienceRevisionPage for user=%v", r.Context().Value("user"))
			err := science.ScienceRevisionPage().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ScienceRevisionPage): %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		})).ServeHTTP(w, r)
	})

	SetupAssetsRoutes(mux)

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

	err = http.ListenAndServe(":"+port, handler)
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
