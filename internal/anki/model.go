package anki

type Deck struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CardCount int    `json:"card_count"`
}

type Card struct {
	ID     string `json:"id"`
	DeckID string `json:"deck_id"`
	Front  string `json:"front"`
	Back   string `json:"back"`
}
