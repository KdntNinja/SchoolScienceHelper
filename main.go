package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/KdntNinja/ScratchClone/assets"
	errorpages "github.com/KdntNinja/ScratchClone/ui/pages/error"
	legalpages "github.com/KdntNinja/ScratchClone/ui/pages/legal"
	publicpages "github.com/KdntNinja/ScratchClone/ui/pages/public"
	userpages "github.com/KdntNinja/ScratchClone/ui/pages/user"
	"github.com/KdntNinja/ScratchClone/utils"
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
		_, _ = fmt.Fprintf(w, `{"AUTH0_DOMAIN":%q,"AUTH0_CLIENT_ID":%q}`, domain, clientID)
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
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		publicpages.Auth().Render(r.Context(), w)
	})
	mux.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		legalpages.Terms().Render(r.Context(), w)
	})
	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		legalpages.Privacy().Render(r.Context(), w)
	})
	mux.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		userpages.Settings().Render(r.Context(), w)
	})

	db := utils.SetupDB()
	utils.SetDB(db)

	mux.HandleFunc("/dash", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		userpages.Dash().Render(r.Context(), w)
	})
	mux.HandleFunc("/api/project/save", utils.HandleProjectSave)
	mux.HandleFunc("/api/project/load", utils.HandleProjectLoad)
	mux.HandleFunc("/api/project/list", utils.HandleProjectList)
	mux.HandleFunc("/api/project/delete", utils.HandleProjectDelete)
	mux.HandleFunc("/api/project/publish", utils.HandleProjectPublish)
	mux.HandleFunc("/api/project/public", utils.HandleProjectLoadPublic)
	// Handler registration for error pages
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
	mux.HandleFunc("/newproject", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		userpages.NewProject().Render(r.Context(), w)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	log.Infof("Server running on :%s", port)
	err := http.ListenAndServe(":"+port, mux)
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
