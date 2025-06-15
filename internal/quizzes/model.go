package quizzes

// Quiz represents a quiz created by a user or the system
type Quiz struct {
	ID        string
	OwnerID   string
	Title     string
	CreatedAt int64
	UpdatedAt int64
}

// Question represents a quiz question
type Question struct {
	ID      string
	QuizID  string
	Prompt  string
	Choices []string
	Answer  int // index of correct answer
}
