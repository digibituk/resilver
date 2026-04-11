package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>BBC News</title>
    <item>
      <title>First headline</title>
      <link>https://example.com/1</link>
    </item>
    <item>
      <title>Second headline</title>
      <link>https://example.com/2</link>
    </item>
    <item>
      <title>Third headline</title>
      <link>https://example.com/3</link>
    </item>
  </channel>
</rss>`

func fakeRSSServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(testRSS))
	}))
}

func TestFetchReturnsHeadlines(t *testing.T) {
	srv := fakeRSSServer(t)
	defer srv.Close()

	client := NewClient()
	items, err := client.Fetch(srv.URL, 5)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
	if items[0].Title != "First headline" {
		t.Errorf("items[0].Title = %q, want First headline", items[0].Title)
	}
	if items[0].Link != "https://example.com/1" {
		t.Errorf("items[0].Link = %q, want https://example.com/1", items[0].Link)
	}
}

func TestFetchRespectsMaxItems(t *testing.T) {
	srv := fakeRSSServer(t)
	defer srv.Close()

	client := NewClient()
	items, err := client.Fetch(srv.URL, 2)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
}

func TestFetchReturnsErrorOnBadResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Fetch(srv.URL, 5)
	if err == nil {
		t.Error("Fetch() expected error on bad response, got nil")
	}
}

func TestFetchReturnsErrorOnInvalidXML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not xml"))
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Fetch(srv.URL, 5)
	if err == nil {
		t.Error("Fetch() expected error on invalid XML, got nil")
	}
}

func TestCachedClientReturnsCachedData(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(testRSS))
	}))
	defer srv.Close()

	cached := NewCachedClient(time.Minute)
	_, err := cached.Fetch(srv.URL, 5)
	if err != nil {
		t.Fatalf("first Fetch() error: %v", err)
	}

	_, err = cached.Fetch(srv.URL, 5)
	if err != nil {
		t.Fatalf("second Fetch() error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("server called %d times, want 1 (cached)", callCount)
	}
}

func TestNewsItemJSON(t *testing.T) {
	item := NewsItem{Title: "Test headline", Link: "https://example.com"}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var got NewsItem
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if got.Title != item.Title {
		t.Errorf("Title = %q, want %q", got.Title, item.Title)
	}
	if got.Link != item.Link {
		t.Errorf("Link = %q, want %q", got.Link, item.Link)
	}
}
