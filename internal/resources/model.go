package resources

type Resource struct {
	ID        string
	OwnerID   string
	Type      string // e.g. 'topic', 'flashcard', 'summary', etc.
	Title     string
	Content   string
	CreatedAt int64
	UpdatedAt int64
}
