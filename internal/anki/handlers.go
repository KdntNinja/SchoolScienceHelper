package anki

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// List all decks for the authenticated user
func ListDecks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(string)
		rows, err := db.Query(`SELECT id, name FROM anki_decks WHERE owner_id = $1`, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var decks []map[string]interface{}
		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err == nil {
				// Count cards in deck
				var cardCount int
				db.QueryRow(`SELECT COUNT(*) FROM anki_cards WHERE deck_id = $1`, id).Scan(&cardCount)
				decks = append(decks, map[string]interface{}{"id": id, "name": name, "card_count": cardCount})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(decks)
	}
}

// List all cards for a given deck (must belong to user)
func ListCards(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(string)
		deckID := r.URL.Query().Get("deck_id")
		if deckID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		rows, err := db.Query(`SELECT id, front, back FROM anki_cards WHERE deck_id = $1 AND owner_id = $2`, deckID, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var cards []map[string]interface{}
		for rows.Next() {
			var id, front, back string
			if err := rows.Scan(&id, &front, &back); err == nil {
				cards = append(cards, map[string]interface{}{"id": id, "front": front, "back": back})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cards)
	}
}

// (Stub) Import handler - expects multipart/form-data with .apkg file
func ImportDeck(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement .apkg parsing and import logic
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":"Not implemented"}`))
	}
}
