package science

import (
	"context"
	"database/sql"
	"log"
)

// CollectAllBoardLinks scrapes and stores links for all boards and kinds
func CollectAllBoardLinks(ctx context.Context, db *sql.DB) {
	for board, kinds := range BoardPages {
		for kind := range kinds {
			log.Printf("Scraping %s %s...", board, kind)
			if err := ScrapeLinks(ctx, db, board, kind); err != nil {
				log.Printf("Error scraping %s %s: %v", board, kind, err)
			}
		}
	}
}
