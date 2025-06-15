package revision

type RevisionResource struct {
	ID        string
	OwnerID   string
	Type      string // flashcard, note, summary, etc.
	Topic     string
	Content   string // markdown or text
	CreatedAt int64
	UpdatedAt int64
}
