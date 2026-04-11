package news

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type NewsItem struct {
	Title  string `json:"title"`
	Link   string `json:"link"`
	Source string `json:"source"`
	Image  string `json:"image,omitempty"`
}

type rssFeed struct {
	Channel struct {
		Title string        `xml:"title"`
		Items []rssFeedItem `xml:"item"`
	} `xml:"channel"`
}

type rssFeedItem struct {
	Title        string        `xml:"title"`
	Link         string        `xml:"link"`
	MediaContent *rssMedia     `xml:"http://search.yahoo.com/mrss/ content"`
	MediaThumb   *rssMedia     `xml:"http://search.yahoo.com/mrss/ thumbnail"`
	Enclosure    *rssEnclosure `xml:"enclosure"`
}

type rssMedia struct {
	URL string `xml:"url,attr"`
}

type rssEnclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Fetch retrieves items from multiple RSS feeds, interleaves them, and
// caps the result at maxItems. Partial failures are tolerated — if at
// least one feed succeeds, items are returned. Returns an error only
// when all feeds fail.
func (c *Client) Fetch(feedURLs []string, maxItems int) ([]NewsItem, error) {
	var allFeeds [][]NewsItem

	for _, url := range feedURLs {
		items, err := c.fetchOne(url)
		if err != nil {
			log.Printf("news feed error (%s): %v", url, err)
			continue
		}
		allFeeds = append(allFeeds, items)
	}

	if len(allFeeds) == 0 {
		return nil, fmt.Errorf("all %d news feeds failed", len(feedURLs))
	}

	merged := interleave(allFeeds)
	if maxItems > 0 && len(merged) > maxItems {
		merged = merged[:maxItems]
	}

	return merged, nil
}

func (c *Client) fetchOne(feedURL string) ([]NewsItem, error) {
	resp, err := c.httpClient.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("news request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("news feed returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read news response: %w", err)
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	items := make([]NewsItem, len(feed.Channel.Items))
	for i, raw := range feed.Channel.Items {
		items[i] = NewsItem{
			Title:  html.UnescapeString(raw.Title),
			Link:   raw.Link,
			Source: html.UnescapeString(feed.Channel.Title),
			Image:  extractImage(raw),
		}
	}

	return items, nil
}

func extractImage(item rssFeedItem) string {
	if item.MediaContent != nil && item.MediaContent.URL != "" {
		return item.MediaContent.URL
	}
	if item.MediaThumb != nil && item.MediaThumb.URL != "" {
		return item.MediaThumb.URL
	}
	if item.Enclosure != nil && item.Enclosure.URL != "" {
		if len(item.Enclosure.Type) >= 5 && item.Enclosure.Type[:5] == "image" {
			return item.Enclosure.URL
		}
	}
	return ""
}

// interleave round-robins items from multiple feeds so sources are evenly
// distributed: [A1, B1, A2, B2, A3, ...].
func interleave(feeds [][]NewsItem) []NewsItem {
	var result []NewsItem
	maxLen := 0
	for _, f := range feeds {
		if len(f) > maxLen {
			maxLen = len(f)
		}
	}
	for i := 0; i < maxLen; i++ {
		for _, f := range feeds {
			if i < len(f) {
				result = append(result, f[i])
			}
		}
	}
	return result
}

type CachedClient struct {
	client *Client
	ttl    time.Duration
	mu     sync.RWMutex
	cached []NewsItem
	expiry time.Time
}

func NewCachedClient(ttl time.Duration) *CachedClient {
	return &CachedClient{
		client: NewClient(),
		ttl:    ttl,
	}
}

func (c *CachedClient) Fetch(feedURLs []string, maxItems int) ([]NewsItem, error) {
	c.mu.RLock()
	if c.cached != nil && time.Now().Before(c.expiry) {
		items := make([]NewsItem, len(c.cached))
		copy(items, c.cached)
		c.mu.RUnlock()
		return items, nil
	}
	c.mu.RUnlock()

	items, err := c.client.Fetch(feedURLs, maxItems)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cached = items
	c.expiry = time.Now().Add(c.ttl)
	c.mu.Unlock()

	return items, nil
}
