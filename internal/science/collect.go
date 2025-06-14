package science

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// CollectAllBoardData scrapes and stores all relevant data for all boards and kinds
func CollectAllBoardData(ctx context.Context, db *sql.DB) {
	for board := range BoardPages {
		log.Printf("[Collect] Scraping specs for %s...", board)
		if err := ScrapeAndStoreSpecs(ctx, db, board); err != nil {
			log.Printf("[Collect] Error scraping specs for %s: %v", board, err)
		} else {
			log.Printf("[Collect] Finished scraping specs for %s", board)
		}
		log.Printf("[Collect] Scraping papers for %s...", board)
		if err := ScrapeAndStorePapers(ctx, db, board); err != nil {
			log.Printf("[Collect] Error scraping papers for %s: %v", board, err)
		} else {
			log.Printf("[Collect] Finished scraping papers for %s", board)
		}
		log.Printf("[Collect] Scraping questions for %s...", board)
		if err := ScrapeAndStoreQuestions(ctx, db, board); err != nil {
			log.Printf("[Collect] Error scraping questions for %s: %v", board, err)
		} else {
			log.Printf("[Collect] Finished scraping questions for %s", board)
		}
		log.Printf("[Collect] Scraping revision for %s...", board)
		if err := ScrapeAndStoreRevision(ctx, db, board); err != nil {
			log.Printf("[Collect] Error scraping revision for %s: %v", board, err)
		} else {
			log.Printf("[Collect] Finished scraping revision for %s", board)
		}
	}
}

// ScrapeAndStoreSpecs fetches and stores all specs for a board
func ScrapeAndStoreSpecs(ctx context.Context, db *sql.DB, board string) error {
	switch board {
	case "aqa":
		return ScrapeAndStoreAqaSpecs(ctx, db)
	case "ocr":
		return ScrapeAndStoreOcrSpecs(ctx, db)
	case "edexcel":
		return ScrapeAndStoreEdexcelSpecs(ctx, db)
	default:
		return nil
	}
}

// ScrapeAndStorePapers fetches and stores all papers for a board
func ScrapeAndStorePapers(ctx context.Context, db *sql.DB, board string) error {
	switch board {
	case "aqa":
		return ScrapeAndStoreAqaPapers(ctx, db)
	case "ocr":
		return ScrapeAndStoreOcrPapers(ctx, db)
	case "edexcel":
		return ScrapeAndStoreEdexcelPapers(ctx, db)
	default:
		return nil
	}
}

// ScrapeAndStoreQuestions fetches and stores all questions for a board
func ScrapeAndStoreQuestions(ctx context.Context, db *sql.DB, board string) error {
	switch board {
	case "aqa":
		return ScrapeAndStoreAqaQuestions(ctx, db)
	case "ocr":
		return ScrapeAndStoreOcrQuestions(ctx, db)
	case "edexcel":
		return ScrapeAndStoreEdexcelQuestions(ctx, db)
	default:
		return nil
	}
}

// ScrapeAndStoreRevision fetches and stores all revision for a board
func ScrapeAndStoreRevision(ctx context.Context, db *sql.DB, board string) error {
	switch board {
	case "aqa":
		return ScrapeAndStoreAqaRevision(ctx, db)
	case "ocr":
		return ScrapeAndStoreOcrRevision(ctx, db)
	case "edexcel":
		return ScrapeAndStoreEdexcelRevision(ctx, db)
	default:
		return nil
	}
}

// --- AQA SPEC SCRAPER ---
func ScrapeAndStoreAqaSpecs(ctx context.Context, db *sql.DB) error {
	url := "https://www.aqa.org.uk/subjects/science/gcse"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	z := html.NewTokenizer(resp.Body)
	var inList bool
	var subject, title, link string
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "ul" {
			for _, a := range tk.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "qual-list") {
					inList = true
				}
			}
		}
		if inList && tk.Type == html.StartTagToken && tk.Data == "a" {
			for _, a := range tk.Attr {
				if a.Key == "href" && strings.Contains(a.Val, "/subjects/science/gcse/") {
					link = "https://www.aqa.org.uk" + a.Val
				}
			}
		}
		if inList && tk.Type == html.TextToken && link != "" {
			title = tk.Data
			if strings.Contains(strings.ToLower(title), "biology") {
				subject = "Biology"
			} else if strings.Contains(strings.ToLower(title), "chemistry") {
				subject = "Chemistry"
			} else if strings.Contains(strings.ToLower(title), "physics") {
				subject = "Physics"
			} else {
				continue
			}
			// Fetch spec content
			specContent, err := fetchAqaSpecContent(link)
			if err != nil {
				continue
			}
			s := Spec{
				Board:   "aqa",
				Tier:    "all",
				Subject: subject,
				Title:   title,
				Content: specContent,
			}
			if err := UpsertSpec(ctx, db, s); err != nil {
				continue
			}
			link = ""
		}
	}
	return nil
}

func fetchAqaSpecContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	var sb strings.Builder
	z := html.NewTokenizer(resp.Body)
	var inContent bool
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "div" {
			for _, a := range tk.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "content") {
					inContent = true
				}
			}
		}
		if inContent && tk.Type == html.TextToken {
			sb.WriteString(tk.Data)
		}
		if inContent && tk.Type == html.EndTagToken && tk.Data == "div" {
			inContent = false
		}
	}
	return sb.String(), nil
}

// --- AQA PAPERS SCRAPER ---
func ScrapeAndStoreAqaPapers(ctx context.Context, db *sql.DB) error {
	url := "https://www.aqa.org.uk/find-past-papers-and-mark-schemes"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	z := html.NewTokenizer(resp.Body)
	var year, subject, paperURL string
	var inTable, inRow bool
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "table" {
			inTable = true
		}
		if inTable && tk.Type == html.StartTagToken && tk.Data == "tr" {
			inRow = true
			year, subject, paperURL = "", "", ""
		}
		if inRow && tk.Type == html.StartTagToken && tk.Data == "a" {
			for _, a := range tk.Attr {
				if a.Key == "href" && strings.Contains(a.Val, ".pdf") {
					paperURL = a.Val
				}
			}
		}
		if inRow && tk.Type == html.TextToken {
			text := strings.TrimSpace(tk.Data)
			if year == "" && len(text) == 4 && strings.HasPrefix(text, "20") {
				year = text
			} else if subject == "" && (strings.Contains(text, "Biology") || strings.Contains(text, "Chemistry") || strings.Contains(text, "Physics")) {
				subject = text
			}
		}
		if inRow && tk.Type == html.EndTagToken && tk.Data == "tr" {
			if year != "" && subject != "" && paperURL != "" {
				p := Paper{
					Board:   "aqa",
					Tier:    "all",
					Year:    parseYear(year),
					Subject: subject,
					URL:     paperURL,
				}
				log.Printf("[Collect] Upserting paper: board=%s, year=%d, subject=%s, url=%s", p.Board, p.Year, p.Subject, p.URL)
				if err := UpsertPaper(ctx, db, p); err != nil {
					log.Printf("[Collect] Error upserting paper: %v", err)
				}
			}
			inRow = false
		}
		if inTable && tk.Type == html.EndTagToken && tk.Data == "table" {
			inTable = false
		}
	}
	return nil
}

func parseYear(s string) int {
	y, _ := strconv.Atoi(s)
	return y
}

// --- AQA QUESTIONS SCRAPER ---
func ScrapeAndStoreAqaQuestions(ctx context.Context, db *sql.DB) error {
	// AQA does not publish a public question bank; this would require scraping from past papers or using a third-party source.
	// For demonstration, we will extract questions from the first page of each paper PDF (not implemented here for brevity).
	// In production, use a PDF parser or integrate with a question API.
	return nil
}

// --- AQA REVISION SCRAPER ---
func ScrapeAndStoreAqaRevision(ctx context.Context, db *sql.DB) error {
	// AQA does not provide official revision notes; this would require scraping from their revision resources or using a third-party source.
	// For demonstration, this is left as a stub.
	return nil
}

// --- OCR SPEC SCRAPER ---
func ScrapeAndStoreOcrSpecs(ctx context.Context, db *sql.DB) error {
	url := "https://www.ocr.org.uk/qualifications/by-subject/science/"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	z := html.NewTokenizer(resp.Body)
	var inList bool
	var subject, title, link string
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "ul" {
			for _, a := range tk.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "qualifications-list") {
					inList = true
				}
			}
		}
		if inList && tk.Type == html.StartTagToken && tk.Data == "a" {
			for _, a := range tk.Attr {
				if a.Key == "href" && strings.Contains(a.Val, "/qualifications/gcse/") {
					link = "https://www.ocr.org.uk" + a.Val
				}
			}
		}
		if inList && tk.Type == html.TextToken && link != "" {
			title = tk.Data
			if strings.Contains(strings.ToLower(title), "biology") {
				subject = "Biology"
			} else if strings.Contains(strings.ToLower(title), "chemistry") {
				subject = "Chemistry"
			} else if strings.Contains(strings.ToLower(title), "physics") {
				subject = "Physics"
			} else {
				continue
			}
			specContent, err := fetchOcrSpecContent(link)
			if err != nil {
				continue
			}
			s := Spec{
				Board:   "ocr",
				Tier:    "all",
				Subject: subject,
				Title:   title,
				Content: specContent,
			}
			if err := UpsertSpec(ctx, db, s); err != nil {
				continue
			}
			link = ""
		}
	}
	return nil
}

func fetchOcrSpecContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	var sb strings.Builder
	z := html.NewTokenizer(resp.Body)
	var inContent bool
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "div" {
			for _, a := range tk.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "content") {
					inContent = true
				}
			}
		}
		if inContent && tk.Type == html.TextToken {
			sb.WriteString(tk.Data)
		}
		if inContent && tk.Type == html.EndTagToken && tk.Data == "div" {
			inContent = false
		}
	}
	return sb.String(), nil
}

// --- OCR PAPERS SCRAPER ---
func ScrapeAndStoreOcrPapers(ctx context.Context, db *sql.DB) error {
	url := "https://www.ocr.org.uk/qualifications/past-papers/"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	z := html.NewTokenizer(resp.Body)
	var year, subject, paperURL string
	var inTable, inRow bool
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "table" {
			inTable = true
		}
		if inTable && tk.Type == html.StartTagToken && tk.Data == "tr" {
			inRow = true
			year, subject, paperURL = "", "", ""
		}
		if inRow && tk.Type == html.StartTagToken && tk.Data == "a" {
			for _, a := range tk.Attr {
				if a.Key == "href" && strings.Contains(a.Val, ".pdf") {
					paperURL = a.Val
				}
			}
		}
		if inRow && tk.Type == html.TextToken {
			text := strings.TrimSpace(tk.Data)
			if year == "" && len(text) == 4 && strings.HasPrefix(text, "20") {
				year = text
			} else if subject == "" && (strings.Contains(text, "Biology") || strings.Contains(text, "Chemistry") || strings.Contains(text, "Physics")) {
				subject = text
			}
		}
		if inRow && tk.Type == html.EndTagToken && tk.Data == "tr" {
			if year != "" && subject != "" && paperURL != "" {
				p := Paper{
					Board:   "ocr",
					Tier:    "all",
					Year:    parseYear(year),
					Subject: subject,
					URL:     paperURL,
				}
				log.Printf("[Collect] Upserting paper: board=%s, year=%d, subject=%s, url=%s", p.Board, p.Year, p.Subject, p.URL)
				if err := UpsertPaper(ctx, db, p); err != nil {
					log.Printf("[Collect] Error upserting paper: %v", err)
				}
			}
			inRow = false
		}
		if inTable && tk.Type == html.EndTagToken && tk.Data == "table" {
			inTable = false
		}
	}
	return nil
}

// --- EDEXCEL SPEC SCRAPER ---
func ScrapeAndStoreEdexcelSpecs(ctx context.Context, db *sql.DB) error {
	url := "https://qualifications.pearson.com/en/qualifications/edexcel-gcses/sciences-2016.html"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	z := html.NewTokenizer(resp.Body)
	var inList bool
	var subject, title, link string
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "ul" {
			for _, a := range tk.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "qualifications-list") {
					inList = true
				}
			}
		}
		if inList && tk.Type == html.StartTagToken && tk.Data == "a" {
			for _, a := range tk.Attr {
				if a.Key == "href" && strings.Contains(a.Val, "/en/qualifications/edexcel-gcses/") {
					link = "https://qualifications.pearson.com" + a.Val
				}
			}
		}
		if inList && tk.Type == html.TextToken && link != "" {
			title = tk.Data
			if strings.Contains(strings.ToLower(title), "biology") {
				subject = "Biology"
			} else if strings.Contains(strings.ToLower(title), "chemistry") {
				subject = "Chemistry"
			} else if strings.Contains(strings.ToLower(title), "physics") {
				subject = "Physics"
			} else {
				continue
			}
			specContent, err := fetchEdexcelSpecContent(link)
			if err != nil {
				continue
			}
			s := Spec{
				Board:   "edexcel",
				Tier:    "all",
				Subject: subject,
				Title:   title,
				Content: specContent,
			}
			if err := UpsertSpec(ctx, db, s); err != nil {
				continue
			}
			link = ""
		}
	}
	return nil
}

func fetchEdexcelSpecContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	var sb strings.Builder
	z := html.NewTokenizer(resp.Body)
	var inContent bool
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "div" {
			for _, a := range tk.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "content") {
					inContent = true
				}
			}
		}
		if inContent && tk.Type == html.TextToken {
			sb.WriteString(tk.Data)
		}
		if inContent && tk.Type == html.EndTagToken && tk.Data == "div" {
			inContent = false
		}
	}
	return sb.String(), nil
}

// --- EDEXCEL PAPERS SCRAPER ---
func ScrapeAndStoreEdexcelPapers(ctx context.Context, db *sql.DB) error {
	url := "https://qualifications.pearson.com/en/support/support-topics/exams/past-papers.html"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	z := html.NewTokenizer(resp.Body)
	var year, subject, paperURL string
	var inTable, inRow bool
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "table" {
			inTable = true
		}
		if inTable && tk.Type == html.StartTagToken && tk.Data == "tr" {
			inRow = true
			year, subject, paperURL = "", "", ""
		}
		if inRow && tk.Type == html.StartTagToken && tk.Data == "a" {
			for _, a := range tk.Attr {
				if a.Key == "href" && strings.Contains(a.Val, ".pdf") {
					paperURL = a.Val
				}
			}
		}
		if inRow && tk.Type == html.TextToken {
			text := strings.TrimSpace(tk.Data)
			if year == "" && len(text) == 4 && strings.HasPrefix(text, "20") {
				year = text
			} else if subject == "" && (strings.Contains(text, "Biology") || strings.Contains(text, "Chemistry") || strings.Contains(text, "Physics")) {
				subject = text
			}
		}
		if inRow && tk.Type == html.EndTagToken && tk.Data == "tr" {
			if year != "" && subject != "" && paperURL != "" {
				p := Paper{
					Board:   "edexcel",
					Tier:    "all",
					Year:    parseYear(year),
					Subject: subject,
					URL:     paperURL,
				}
				log.Printf("[Collect] Upserting paper: board=%s, year=%d, subject=%s, url=%s", p.Board, p.Year, p.Subject, p.URL)
				if err := UpsertPaper(ctx, db, p); err != nil {
					log.Printf("[Collect] Error upserting paper: %v", err)
				}
			}
			inRow = false
		}
		if inTable && tk.Type == html.EndTagToken && tk.Data == "table" {
			inTable = false
		}
	}
	return nil
}

// --- EDEXCEL QUESTIONS SCRAPER ---
func ScrapeAndStoreEdexcelQuestions(ctx context.Context, db *sql.DB) error {
	// Edexcel does not publish a public question bank; this would require scraping from past papers or using a third-party source.
	return nil
}

// --- EDEXCEL REVISION SCRAPER ---
func ScrapeAndStoreEdexcelRevision(ctx context.Context, db *sql.DB) error {
	// Edexcel does not provide official revision notes; this would require scraping from their revision resources or using a third-party source.
	return nil
}

// --- OCR QUESTIONS SCRAPER ---
func ScrapeAndStoreOcrQuestions(ctx context.Context, db *sql.DB) error {
	// OCR does not publish a public question bank; this would require scraping from past papers or using a third-party source.
	return nil
}

// --- OCR REVISION SCRAPER ---
func ScrapeAndStoreOcrRevision(ctx context.Context, db *sql.DB) error {
	// OCR does not provide official revision notes; this would require scraping from their revision resources or using a third-party source.
	return nil
}
