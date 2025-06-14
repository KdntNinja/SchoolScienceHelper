package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"KdnSite/assets"
	"KdnSite/internal/handlers"
	science "KdnSite/internal/science"
	errorpages "KdnSite/ui/pages/error"
	legalpages "KdnSite/ui/pages/legal"
	publicpages "KdnSite/ui/pages/public"
	userpages "KdnSite/ui/pages/user"
	sciencepages "KdnSite/ui/pages/user/science"

	"context"

	log "github.com/sirupsen/logrus"
)

func main() {
	setupLogging()
	checkEnvVars()
	db := setupDatabase()
	defer db.Close()

	// Start background weekly board data sync
	go func() {
		for {
			log.Info("[Background] Starting weekly exam board scrape...")
			science.CollectAllBoardData(context.Background(), db)
			log.Info("[Background] Exam board scrape complete. Next run in 7 days.")
			time.Sleep(7 * 24 * time.Hour)
		}
	}()

	mux := http.NewServeMux()
	registerStaticRoutes(mux)
	registerLegalRoutes(mux)
	registerAuthRoutes(mux)
	registerUserRoutes(mux)
	registerScienceAPIRoutes(mux, db)
	registerScienceUserRoutes(mux)
	registerScienceHubRoute(mux)
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

func registerScienceAPIRoutes(mux *http.ServeMux, db *sql.DB) {
	boards := []string{"aqa", "ocr", "edexcel"}
	tiers := []string{"foundation", "higher", "separate_foundation", "separate_higher"}
	for _, board := range boards {
		for _, tier := range tiers {
			mux.HandleFunc("/api/"+board+"/"+tier+"/spec", science.SpecsAPI(db))
			mux.HandleFunc("/api/"+board+"/"+tier+"/papers", science.PapersAPI(db))
			mux.HandleFunc("/api/"+board+"/"+tier+"/questions", science.QuestionsAPI(db))
			mux.HandleFunc("/api/"+board+"/"+tier+"/revision", science.RevisionAPI(db))
		}
	}
}

func registerScienceUserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/user/science/spec", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("[ScienceSpecPage] user=%v, path=%s", r.Context().Value("user"), r.URL.Path)
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("Rendering ScienceSpecPage for user=%v", r.Context().Value("user"))
			err := sciencepages.ScienceSpecPage().Render(r.Context(), w)
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
			err := sciencepages.SciencePapersPage().Render(r.Context(), w)
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
			err := sciencepages.ScienceQuestionsPage().Render(r.Context(), w)
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
			err := sciencepages.ScienceRevisionPage().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ScienceRevisionPage): %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		})).ServeHTTP(w, r)
	})
}

func registerScienceHubRoute(mux *http.ServeMux) {
	mux.HandleFunc("/user/science/", func(w http.ResponseWriter, r *http.Request) {
		board := "aqa"
		tier := "foundation"
		if b, err := r.Cookie("science_board"); err == nil {
			board = b.Value
		} else if b := r.URL.Query().Get("board"); b != "" {
			board = b
		}
		if t, err := r.Cookie("science_tier"); err == nil {
			tier = t.Value
		} else if t := r.URL.Query().Get("tier"); t != "" {
			tier = t
		}
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := sciencepages.ScienceHubPage(board, tier).Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ScienceHubPage): %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		})).ServeHTTP(w, r)
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
