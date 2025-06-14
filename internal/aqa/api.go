package aqa

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// URL pattern: /api/{board}/{tier}/spec, etc.

func SpecsAPI(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid URL: expected /api/{board}/{tier}/spec"))
			return
		}
		board := parts[1]
		tier := parts[2]
		specs, err := GetSpecsByBoardTier(r.Context(), db, board, tier)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error fetching specs"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(specs)
	}
}

func PapersAPI(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid URL: expected /api/{board}/{tier}/papers"))
			return
		}
		board := parts[1]
		tier := parts[2]
		papers, err := GetPapersByBoardTier(r.Context(), db, board, tier)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error fetching papers"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(papers)
	}
}

func QuestionsAPI(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid URL: expected /api/{board}/{tier}/questions"))
			return
		}
		board := parts[1]
		tier := parts[2]
		questions, err := GetQuestionsByBoardTier(r.Context(), db, board, tier)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error fetching questions"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(questions)
	}
}

func RevisionAPI(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid URL: expected /api/{board}/{tier}/revision"))
			return
		}
		board := parts[1]
		tier := parts[2]
		revs, err := GetRevisionByBoardTier(r.Context(), db, board, tier)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error fetching revision resources"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(revs)
	}
}
