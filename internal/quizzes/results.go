package quizzes

type QuizResult struct {
	ID        string
	QuizID    string
	UserID    string
	Score     int
	StartedAt int64
	EndedAt   int64
	Answers   []int // index of selected answers
}
