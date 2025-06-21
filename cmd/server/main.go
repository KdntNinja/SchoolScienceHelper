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
	"KdnSite/internal/quiz"
	"KdnSite/internal/resources"
	"KdnSite/internal/revision"
	"KdnSite/internal/user"
	adminpages "KdnSite/ui/pages/admin"
	errorpages "KdnSite/ui/pages/error"
	legalpages "KdnSite/ui/pages/legal"
	publicpages "KdnSite/ui/pages/public"
	userpages "KdnSite/ui/pages/user"
	userpagescommunity "KdnSite/ui/pages/user/community"
	userpagesprojects "KdnSite/ui/pages/user/projects"
	userpagesquiz "KdnSite/ui/pages/user/quiz"

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
	dbURL := os.Getenv("POSTGRES_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("POSTGRES_DATABASE_URL environment variable is not set!")
	}
}

func setupDatabase() *sql.DB {
	dbURL := os.Getenv("POSTGRES_DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres DB: %v", err)
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
	mux.HandleFunc("/api/auth/callback", handlers.HandleAuthCallback)
	mux.HandleFunc("/api/auth/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/auth/delete", handlers.DeleteAccountHandler)
	mux.HandleFunc("/api/auth/change-password", handlers.ChangePasswordHandler)
	mux.HandleFunc("/api/auth/change-username", handlers.ChangeUsernameHandler)
	mux.HandleFunc("/api/auth/avatar", handlers.AvatarUploadHandler)
	mux.Handle("/api/auth/resend-verification", handlers.RequireAuth(http.HandlerFunc(handlers.ResendVerificationHandler)))
	mux.Handle("/api/auth/logout-all", handlers.RequireAuth(http.HandlerFunc(handlers.LogoutAllHandler)))
	mux.Handle("/api/auth/sessions", handlers.RequireAuth(http.HandlerFunc(handlers.SessionsHandler)))
	mux.HandleFunc("/api/auth/check", handlers.AuthCheckHandler)
	mux.HandleFunc("/api/auth/current-username", func(w http.ResponseWriter, r *http.Request) {
		username, err := handlers.GetUsernameFromJWT(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(""))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(username))
	})
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
			err := userpagesprojects.ProjectList().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ProjectList): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/projects/editor", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpagesprojects.ProjectEditor().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (ProjectEditor): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/community/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpagescommunity.Leaderboard().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (Leaderboard): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/settings", func(w http.ResponseWriter, r *http.Request) {
		domain := os.Getenv("AUTH0_DOMAIN")
		clientID := os.Getenv("AUTH0_CLIENT_ID")
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpages.Settings(domain, clientID).Render(r.Context(), w)
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
	mux.HandleFunc("/error/verifyemail", func(w http.ResponseWriter, r *http.Request) {
		err := errorpages.VerifyEmail().Render(r.Context(), w)
		if err != nil {
			log.Errorf("Render error (VerifyEmail): %v", err)
		}
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
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value("user")
			var claims map[string]interface{}
			if user != nil {
				if c, ok := user.(map[string]interface{}); ok {
					claims = c
				}
			}
			adminPerm := os.Getenv("ADMIN_PERMISSION")
			if claims == nil || adminPerm == "" || !hasPermission(claims, adminPerm) {
				w.WriteHeader(http.StatusForbidden)
				_ = errorpages.Forbidden().Render(r.Context(), w)
				return
			}
			_ = adminpages.AdminPanel().Render(r.Context(), w)
		})).ServeHTTP(w, r)
	})

	mux.HandleFunc("/user/quizzes", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpagesquiz.QuizList().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (QuizList): %v", err)
			}
		})).ServeHTTP(w, r)
	})
	mux.HandleFunc("/user/quiz/take", func(w http.ResponseWriter, r *http.Request) {
		handlers.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := userpagesquiz.TakeQuiz().Render(r.Context(), w)
			if err != nil {
				log.Errorf("Render error (TakeQuiz): %v", err)
			}
		})).ServeHTTP(w, r)
	})
}

func SetupAssetsRoutes(mux *http.ServeMux) {
	var isDevelopment = os.Getenv("GO_ENV") != "production"
	var fs http.Handler
	if isDevelopment {
		fs = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			http.FileServer(http.Dir("./assets")).ServeHTTP(w, r)
		})
	} else {
		fs = http.FileServer(http.FS(assets.Assets))
	}
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
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
	mux.HandleFunc("/api/sse/email-verified", handlers.EmailVerificationSSE)
	mux.Handle("/api/quizzes", handlers.RequireAuth(quiz.ListQuizzes(db)))
	mux.Handle("/api/quiz", handlers.RequireAuth(quiz.GetQuiz(db)))
	mux.Handle("/api/quiz/attempt", handlers.RequireAuth(quiz.SubmitQuizAttempt(db)))
}

func hasPermission(claims map[string]interface{}, permission string) bool {
	perms, ok := claims["permissions"].([]interface{})
	if !ok {
		return false
	}
	for _, p := range perms {
		if pstr, ok := p.(string); ok && pstr == permission {
			return true
		}
	}
	return false
}
