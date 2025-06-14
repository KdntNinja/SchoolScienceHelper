package science

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// BoardPage holds info for scraping
var BoardPages = map[string]map[string]string{
	"aqa": {
		"papers": "https://www.aqa.org.uk/find-past-papers-and-mark-schemes",
		"spec":   "https://www.aqa.org.uk/subjects/science/gcse",
	},
	"ocr": {
		"papers": "https://www.ocr.org.uk/qualifications/past-papers/",
		"spec":   "https://www.ocr.org.uk/qualifications/by-subject/science/",
	},
	"edexcel": {
		"papers": "https://qualifications.pearson.com/en/support/support-topics/exams/past-papers.html",
		"spec":   "https://qualifications.pearson.com/en/qualifications/edexcel-gcses/sciences-2016.html",
	},
}

// ScrapeLinks scrapes all links from a board's page and stores them in the DB
func ScrapeLinks(ctx context.Context, db *sql.DB, board, kind string) error {
	url, ok := BoardPages[board][kind]
	if !ok {
		return errors.New("no url for board/kind")
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	z := html.NewTokenizer(resp.Body)
	var links []string
	for {
		t := z.Next()
		if t == html.ErrorToken {
			break
		}
		tk := z.Token()
		if tk.Type == html.StartTagToken && tk.Data == "a" {
			for _, a := range tk.Attr {
				if a.Key == "href" && strings.HasPrefix(a.Val, "http") {
					links = append(links, a.Val)
				}
			}
		}
	}
	for _, l := range links {
		_, err := db.ExecContext(ctx, `INSERT INTO board_links (board, kind, url) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, board, kind, l)
		if err != nil {
			return err
		}
	}
	return nil
}
