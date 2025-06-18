package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/r3labs/sse/v2"
)

var sseServer = sse.New()

func init() {
	sseServer.AutoReplay = false
}

// SSE endpoint for email verification status
type EmailStatus struct {
	Verified bool `json:"verified"`
}

func EmailVerificationSSE(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromJWT(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ch := make(chan bool, 1)
	// Simulate polling for demo; in production, trigger on actual verification event
	go func() {
		for {
			verified := checkEmailVerified(userID)
			if verified {
				ch <- true
				return
			}
			time.Sleep(3 * time.Second)
		}
	}()
	sseServer.CreateStream(userID)
	defer sseServer.RemoveStream(userID)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			enc.Encode(EmailStatus{Verified: true})
			flusher.Flush()
			return
		}
	}
}

// Dummy function, replace with real DB/Auth0 check
func checkEmailVerified(userID string) bool {
	// TODO: Implement real check
	return false
}
