package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"KdnSite/assets"
	"KdnSite/internal/achievements"
	"KdnSite/internal/handlers"
	"KdnSite/internal/leaderboard"
	"KdnSite/internal/projects"
	"KdnSite/internal/quizzes"
	"KdnSite/internal/resources"
	"KdnSite/internal/revision"
	"KdnSite/internal/user"
	errorpages "KdnSite/ui/pages/error"
	legalpages "KdnSite/ui/pages/legal"
	publicpages "KdnSite/ui/pages/public"
	userpages "KdnSite/ui/pages/user"
	userpages_community "KdnSite/ui/pages/user/community"
	userpages_projects "KdnSite/ui/pages/user/projects"
	userpages_quizzes "KdnSite/ui/pages/user/quizzes"

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
	registerAPIRoutes(mux, db)

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
		handlers.RequireAuth(http.HandlerFunc(handlers.DashPageHandler)).ServeHTTP(w, r)
	})

	mux.HandleFunc("/user/projects/list", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages_projects.ProjectList().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ProjectList): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/projects/editor", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages_projects.ProjectEditor().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ProjectEditor): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/quizzes/list", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages_quizzes.QuizList().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (QuizList): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/quizzes/editor", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages_quizzes.QuizEditor().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (QuizEditor): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/community/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages_community.Leaderboard().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (Leaderboard): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/settings", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages.Settings().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (Settings): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/revision", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages.Revision().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (Revision): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/quizzes/quizflow", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages_quizzes.QuizFlow().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (QuizFlow): %v", err)
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

func registerAPIRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.Handle("/api/projects", handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projects.ListProjects(db)(w, r)
		case http.MethodPost:
			projects.CreateProject(db)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/quizzes", handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			quizzes.ListQuizzes(db)(w, r)
		case http.MethodPost:
			quizzes.CreateQuiz(db)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/revision", handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			revision.ListRevisionResources(db)(w, r)
		case http.MethodPost:
			revision.CreateRevisionResource(db)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/leaderboard", handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			leaderboard.ListLeaderboard(db)(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/achievements", handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			achievements.ListAchievements(db)(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/user/profile", handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			user.GetProfile(db)(w, r)
		case http.MethodPost:
			user.UpdateProfile(db)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/resources", handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resources.ListResources(db)(w, r)
		case http.MethodPost:
			resources.CreateResource(db)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
}
