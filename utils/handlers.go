package utils

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandleHealthCheck] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandlePing(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HandlePing] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
