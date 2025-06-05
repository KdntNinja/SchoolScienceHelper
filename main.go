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
	mux.Handle("GET /dash", templ.Handler(pages.Dash()))
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
		_ = utils.SetSession(w, username)
		http.Redirect(w, r, "/dash", http.StatusSeeOther)
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
		err := utils.CreateUser(r.Context(), email, password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not create user"))
			return
		}
		_ = utils.SetSession(w, email)
		http.Redirect(w, r, "/dash", http.StatusSeeOther)
	}))

	mux.Handle("POST /logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.ClearSession(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

	mux.Handle("GET /terms", templ.Handler(pages.Terms()))
	mux.Handle("GET /privacy", templ.Handler(pages.Privacy()))
	mux.Handle("GET /settings", templ.Handler(pages.Settings()))
	mux.Handle("POST /settings", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid form"))
			return
		}
		username, err := utils.GetSessionUsername(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not logged in"))
			return
		}
		newUsername := r.FormValue("username")
		newEmail := r.FormValue("email")
		newPassword := r.FormValue("password")
		err = utils.UpdateUser(r.Context(), username, newUsername, newEmail, newPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not update user: " + err.Error()))
			return
		}
		if newUsername != "" && newUsername != username {
			_ = utils.SetSession(w, newUsername)
		}
		w.Write([]byte("Settings updated!"))
	}))
	mux.Handle("POST /settings/delete", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, err := utils.GetSessionUsername(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not logged in"))
			return
		}
		err = utils.DeleteUser(r.Context(), username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not delete user: " + err.Error()))
			return
		}
		utils.ClearSession(w)
		w.Write([]byte("Account deleted"))
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
