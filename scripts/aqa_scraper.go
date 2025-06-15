package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Placeholder: implement real scraping logic for AQA or other exam boards
func fetchQuestions() []Question {
	// TODO: Scrape or download questions from AQA
	return []Question{}
}

func fetchMarkSchemes() []MarkScheme {
	// TODO: Scrape/download mark schemes
	return []MarkScheme{}
}

func fetchPastPapers() []PastPaper {
	// TODO: Scrape/download past papers
	return []PastPaper{}
}

type Question struct {
	ID          string
	QuizID      string
	Subject     string
	Topic       string
	Prompt      string
	Choices     []string
	Answer      int
	Explanation string
}

type MarkScheme struct {
	ID      string
	PaperID string
	Content string
}

type PastPaper struct {
	ID      string
	Subject string
	Year    string
	URL     string
}

func main() {
	dbURL := os.Getenv("POSTGRES_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("POSTGRES_DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Example: run every 24 hours (or call from cron)
	for {
		fmt.Println("[Scraper] Fetching new data...")
		questions := fetchQuestions()
		for _, q := range questions {
			_, err := db.Exec(`INSERT INTO questions (id, quiz_id, subject, topic, prompt, choices, answer, explanation) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (id) DO NOTHING`, q.ID, q.QuizID, q.Subject, q.Topic, q.Prompt, pqArray(q.Choices), q.Answer, q.Explanation)
			if err != nil {
				log.Println("Error inserting question:", err)
			}
		}
		// TODO: Insert mark schemes and past papers similarly
		fmt.Println("[Scraper] Done. Sleeping 24h...")
		time.Sleep(24 * time.Hour)
	}
}

// pqArray is a helper for pq.Array without importing pq in this snippet
func pqArray(a []string) interface{} { return a }
