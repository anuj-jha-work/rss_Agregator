package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	database "github.com/anuj-jha-work/rss_Agregator/internal/database"
	"github.com/google/uuid"
)

func startScraping(
	db *database.Queries,
	concurency int,
	timebetweenRequests time.Duration,
) {
	log.Printf("Starting scraping with concurrency: %d, time between requests: %s", concurency, timebetweenRequests)
	ticker := time.NewTicker(timebetweenRequests)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurency),
		)
		if err != nil {
			log.Printf("Error fetching next feeds to fetch: %v", err)
			continue
		}
		wg := sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)
			go scrapeFeed(db, &wg, feed)
		}

		wg.Wait()
	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()
	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Error marking feed as fetched: %v", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Printf("Error fetching RSS feed from URL %s: %v", feed.Url, err)
		return
	}
	for _, item := range rssFeed.Channel.Item {
		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}
		publishedAt, err := parsePublicationDate(item.PubDate)
		if err != nil {
			log.Printf("Error parsing publication date %s: %v", item.PubDate, err)
			continue
		}

		_, err = db.CreatePost(context.Background(),
			database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Title:       item.Title,
				Description: description,
				PublishedAt: publishedAt,
				URL:         item.Link,
				FeedID:      feed.ID,
			})
		if err != nil {
			log.Printf("Error creating post: %v", err)
			continue
		}
	}
	log.Printf("Feed %s collected ,%v post Found ", feed.Name, len(rssFeed.Channel.Item))

}

func parsePublicationDate(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC1123, time.RFC1123Z, time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported publication date format")
}
