package main

import (
	"context"
	"log"
	"sync"
	"time"

	database "github.com/anuj-jha-work/rss_Agregator/internal/database"
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
		log.Printf("Found Post: %s on Feed: %s %s", item.Title, feed.Url, feed.Name)
	}
	log.Printf("Feed %s collected ,%v post Found ", feed.Name, len(rssFeed.Channel.Item))

}
