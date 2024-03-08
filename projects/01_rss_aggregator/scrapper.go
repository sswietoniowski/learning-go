package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/sswietoniowski/learning-go/projects/01_rss_aggregator/internal/database"
)

func startScrapping(db *database.Queries, concurrency int, timeBetweenRequests time.Duration) {
	log.Printf("Collecting feeds every %s on %v goroutines...", timeBetweenRequests, concurrency)
	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Printf("Couldn't get next feeds to fetch: %v", err)
			continue
		}
		log.Printf("Found %d feeds to fetch", len(feeds))

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)
			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()
	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s as fetched: %v", feed.Name, err)
		return
	}

	rssFeed, err := fetchFeed(feed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v", feed.Url, err)
		return
	}

	for _, item := range rssFeed.Channel.Items {
		log.Printf("Found post: %s", item.Title)
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(rssFeed.Channel.Items))
}

type RssFeed struct {
	Channel RssChannel `xml:"channel"`
}

type RssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	Items       []RssItem `xml:"item"`
}

type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(feedUrl string) (*RssFeed, error) {
	const timeout = 10 * time.Second
	httpClient := http.Client{
		Timeout: timeout,
	}
	resp, err := httpClient.Get(feedUrl)
	if err != nil {
		log.Printf("Error fetching feed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching feed: status code %d", resp.StatusCode)
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	var rssFeed RssFeed
	err = xml.Unmarshal(body, &rssFeed)
	if err != nil {
		log.Printf("Error unmarshalling XML: %v", err)
		return nil, err
	}

	return &rssFeed, nil
}
