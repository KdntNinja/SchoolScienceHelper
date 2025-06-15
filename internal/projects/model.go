package projects

// Project represents a user-created project (visual program, etc.)
type Project struct {
	ID        string
	OwnerID   string
	Title     string
	CreatedAt int64
	UpdatedAt int64
	Data      string // JSON or other serialized format
}
