package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: import_aqa_questions <csvfile> <DATABASE_URL>")
		os.Exit(1)
	}
	csvfile := os.Args[1]
	dburl := os.Args[2]
	f, err := os.Open(csvfile)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(f)
	db, err := sql.Open("postgres", dburl)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	for i, rec := range records {
		if i == 0 { // skip header
			continue
		}
		id := rec[0]
		quizID := rec[1]
		subject := rec[2]
		topic := rec[3]
		prompt := rec[4]
		choices := strings.Split(rec[5], "|")
		answer := rec[6]
		// Convert answer to int for DB
		answerInt := 0
		fmt.Sscanf(answer, "%d", &answerInt)
		explanation := rec[7]
		_, err := db.Exec(`INSERT INTO questions (id, quiz_id, subject, topic, prompt, choices, answer, explanation) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (id) DO NOTHING`, id, quizID, subject, topic, prompt, pq.Array(choices), answerInt, explanation)
		if err != nil {
			fmt.Printf("Error on row %d: %v\n", i, err)
		}
	}
	fmt.Println("Import complete.")
}
