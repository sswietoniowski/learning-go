package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

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
