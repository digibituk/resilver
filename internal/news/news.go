package news

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type NewsItem struct {
	Title string `json:"title" xml:"title"`
	Link  string `json:"link" xml:"link"`
}

type rssFeed struct {
	Channel struct {
		Items []NewsItem `xml:"item"`
	} `xml:"channel"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Fetch(feedURL string, maxItems int) ([]NewsItem, error) {
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

	items := feed.Channel.Items
	if maxItems > 0 && len(items) > maxItems {
		items = items[:maxItems]
	}

	return items, nil
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

func (c *CachedClient) Fetch(feedURL string, maxItems int) ([]NewsItem, error) {
	c.mu.RLock()
	if c.cached != nil && time.Now().Before(c.expiry) {
		items := make([]NewsItem, len(c.cached))
		copy(items, c.cached)
		c.mu.RUnlock()
		return items, nil
	}
	c.mu.RUnlock()

	items, err := c.client.Fetch(feedURL, maxItems)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cached = items
	c.expiry = time.Now().Add(c.ttl)
	c.mu.Unlock()

	return items, nil
}
