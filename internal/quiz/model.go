package quiz

type Quiz struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Topic       string     `json:"topic"`
	Questions   []Question `json:"questions"`
}

type Question struct {
	ID          string   `json:"id"`
	QuizID      string   `json:"quiz_id"`
	Prompt      string   `json:"prompt"`
	Options     []string `json:"options"`
	Answer      int      `json:"answer"` // index of correct option
	Explanation string   `json:"explanation"`
	Difficulty  string   `json:"difficulty"`
}

type UserQuizAttempt struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	QuizID    string `json:"quiz_id"`
	Answers   []int  `json:"answers"`
	Score     int    `json:"score"`
	Timestamp string `json:"timestamp"`
}
