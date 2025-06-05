package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/KdntNinja/ScratchClone/assets"
	"github.com/KdntNinja/ScratchClone/ui/pages"
	"github.com/KdntNinja/ScratchClone/utils"
	"github.com/a-h/templ"
)

func main() {
	utils.InitDB()
	mux := http.NewServeMux()
	SetupAssetsRoutes(mux)
	mux.Handle("GET /", templ.Handler(pages.Landing()))
	mux.Handle("GET /auth", templ.Handler(pages.Auth()))
	mux.Handle("POST /login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid form"))
			return
		}
		username := r.FormValue("email")
		password := r.FormValue("password")
		ok, err := utils.AuthenticateUser(r.Context(), username, password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server error"))
			return
		}
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid credentials"))
			return
		}
		w.Write([]byte("Login successful!"))
	}))

	mux.Handle("POST /signup", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid form"))
			return
		}
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirm := r.FormValue("confirmPassword")
		if password != confirm {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Passwords do not match"))
			return
		}
		// For now, use email as username in DB
		err := utils.CreateUser(r.Context(), email, password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not create user"))
			return
		}
		w.Write([]byte("Signup successful!"))
	}))

	fmt.Println("Server is running on http://localhost:8090")
	err := http.ListenAndServe(":8090", mux)
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
