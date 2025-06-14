package science

// GCSE Spec
// Subject: e.g. "Biology", "Chemistry", "Physics"
type Spec struct {
	ID      int    `json:"id"`
	Board   string `json:"board"` // aqa, ocr, edexcel, etc.
	Tier    string `json:"tier"`  // foundation, higher, separated_foundation, separated_higher
	Subject string `json:"subject"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Paper struct {
	ID      int    `json:"id"`
	Board   string `json:"board"`
	Tier    string `json:"tier"`
	Year    int    `json:"year"`
	Subject string `json:"subject"`
	URL     string `json:"url"`
}

type Question struct {
	ID       int    `json:"id"`
	Board    string `json:"board"`
	Tier     string `json:"tier"`
	Subject  string `json:"subject"`
	Topic    string `json:"topic"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type Revision struct {
	ID      int    `json:"id"`
	Board   string `json:"board"`
	Tier    string `json:"tier"`
	Subject string `json:"subject"`
	Topic   string `json:"topic"`
	Content string `json:"content"`
}
