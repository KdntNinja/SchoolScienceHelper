package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/KdntNinja/ScratchClone/assets"
	"github.com/KdntNinja/ScratchClone/ui/pages"
	"github.com/KdntNinja/ScratchClone/utils"
	"github.com/a-h/templ"
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
		templ.Handler(pages.Landing()).ServeHTTP(w, r)
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		templ.Handler(pages.Auth()).ServeHTTP(w, r)
	})
	mux.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		templ.Handler(pages.Terms()).ServeHTTP(w, r)
	})
	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		templ.Handler(pages.Privacy()).ServeHTTP(w, r)
	})

	db := utils.SetupDB()
	utils.SetDB(db)

	mux.HandleFunc("/dash", handleDash)
	mux.HandleFunc("/api/project/save", utils.HandleProjectSave)
	mux.HandleFunc("/api/project/load", utils.HandleProjectLoad)
	mux.HandleFunc("/api/project/list", utils.HandleProjectList)
	mux.HandleFunc("/api/project/delete", utils.HandleProjectDelete)
	mux.HandleFunc("/api/project/publish", utils.HandleProjectPublish)
	mux.HandleFunc("/api/project/public", utils.HandleProjectLoadPublic)

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

func handleDash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// No need to fetch projects here; frontend loads them via /api/project/list
	if err := pages.Dash().Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
	}
}
