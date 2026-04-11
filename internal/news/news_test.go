package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testRSSA = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:media="http://search.yahoo.com/mrss/">
  <channel>
    <title>Feed A</title>
    <item>
      <title>A first</title>
      <link>https://a.com/1</link>
      <media:content url="https://a.com/img1.jpg" medium="image" />
    </item>
    <item>
      <title>A second</title>
      <link>https://a.com/2</link>
      <media:thumbnail url="https://a.com/thumb2.jpg" />
    </item>
    <item>
      <title>A third</title>
      <link>https://a.com/3</link>
    </item>
  </channel>
</rss>`

const testRSSB = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Feed B</title>
    <item>
      <title>B first</title>
      <link>https://b.com/1</link>
      <enclosure url="https://b.com/img1.jpg" type="image/jpeg" />
    </item>
    <item>
      <title>B second</title>
      <link>https://b.com/2</link>
    </item>
  </channel>
</rss>`

const testRSSHtmlEntities = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Feed &amp; Source</title>
    <item>
      <title>Tom &amp; Jerry&#39;s &quot;adventure&quot;</title>
      <link>https://example.com/1</link>
    </item>
  </channel>
</rss>`

func fakeRSSServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(body))
	}))
}

func TestFetchSingleFeed(t *testing.T) {
	srv := fakeRSSServer(t, testRSSA)
	defer srv.Close()

	client := NewClient()
	items, err := client.Fetch([]string{srv.URL}, 5)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
	if items[0].Title != "A first" {
		t.Errorf("items[0].Title = %q, want A first", items[0].Title)
	}
	if items[0].Link != "https://a.com/1" {
		t.Errorf("items[0].Link = %q, want https://a.com/1", items[0].Link)
	}
	if items[0].Source != "Feed A" {
		t.Errorf("items[0].Source = %q, want Feed A", items[0].Source)
	}
	if items[0].Image != "https://a.com/img1.jpg" {
		t.Errorf("items[0].Image = %q, want https://a.com/img1.jpg", items[0].Image)
	}
	if items[1].Image != "https://a.com/thumb2.jpg" {
		t.Errorf("items[1].Image = %q, want https://a.com/thumb2.jpg (media:thumbnail)", items[1].Image)
	}
	if items[2].Image != "" {
		t.Errorf("items[2].Image = %q, want empty (no image)", items[2].Image)
	}
}

func TestFetchMultipleFeeds(t *testing.T) {
	srvA := fakeRSSServer(t, testRSSA)
	defer srvA.Close()
	srvB := fakeRSSServer(t, testRSSB)
	defer srvB.Close()

	client := NewClient()
	items, err := client.Fetch([]string{srvA.URL, srvB.URL}, 20)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if len(items) != 5 {
		t.Fatalf("got %d items, want 5", len(items))
	}

	// Should interleave: A1, B1, A2, B2, A3
	if items[0].Title != "A first" {
		t.Errorf("items[0].Title = %q, want A first", items[0].Title)
	}
	if items[0].Source != "Feed A" {
		t.Errorf("items[0].Source = %q, want Feed A", items[0].Source)
	}
	if items[1].Title != "B first" {
		t.Errorf("items[1].Title = %q, want B first", items[1].Title)
	}
	if items[1].Source != "Feed B" {
		t.Errorf("items[1].Source = %q, want Feed B", items[1].Source)
	}
	if items[1].Image != "https://b.com/img1.jpg" {
		t.Errorf("items[1].Image = %q, want https://b.com/img1.jpg (enclosure)", items[1].Image)
	}
	if items[2].Title != "A second" {
		t.Errorf("items[2].Title = %q, want A second", items[2].Title)
	}
	if items[3].Title != "B second" {
		t.Errorf("items[3].Title = %q, want B second", items[3].Title)
	}
	if items[4].Title != "A third" {
		t.Errorf("items[4].Title = %q, want A third", items[4].Title)
	}
}

func TestFetchMultiFeedRespectsMaxItems(t *testing.T) {
	srvA := fakeRSSServer(t, testRSSA)
	defer srvA.Close()
	srvB := fakeRSSServer(t, testRSSB)
	defer srvB.Close()

	client := NewClient()
	items, err := client.Fetch([]string{srvA.URL, srvB.URL}, 3)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
}

func TestFetchPartialFailureStillReturnsResults(t *testing.T) {
	srvA := fakeRSSServer(t, testRSSA)
	defer srvA.Close()
	srvBad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srvBad.Close()

	client := NewClient()
	items, err := client.Fetch([]string{srvA.URL, srvBad.URL}, 20)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3 (from feed A only)", len(items))
	}
}

func TestFetchAllFeedsFailReturnsError(t *testing.T) {
	srvBad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srvBad.Close()

	client := NewClient()
	_, err := client.Fetch([]string{srvBad.URL}, 5)
	if err == nil {
		t.Error("Fetch() expected error when all feeds fail, got nil")
	}
}

func TestFetchReturnsErrorOnInvalidXML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not xml"))
	}))
	defer srv.Close()

	client := NewClient()
	_, err := client.Fetch([]string{srv.URL}, 5)
	if err == nil {
		t.Error("Fetch() expected error on invalid XML, got nil")
	}
}

func TestFetchDecodesHtmlEntities(t *testing.T) {
	srv := fakeRSSServer(t, testRSSHtmlEntities)
	defer srv.Close()

	client := NewClient()
	items, err := client.Fetch([]string{srv.URL}, 5)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	want := `Tom & Jerry's "adventure"`
	if items[0].Title != want {
		t.Errorf("items[0].Title = %q, want %q", items[0].Title, want)
	}
	if items[0].Source != "Feed & Source" {
		t.Errorf("items[0].Source = %q, want Feed & Source", items[0].Source)
	}
}

func TestCachedClientReturnsCachedData(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(testRSSA))
	}))
	defer srv.Close()

	cached := NewCachedClient(time.Minute)
	_, err := cached.Fetch([]string{srv.URL}, 5)
	if err != nil {
		t.Fatalf("first Fetch() error: %v", err)
	}

	_, err = cached.Fetch([]string{srv.URL}, 5)
	if err != nil {
		t.Fatalf("second Fetch() error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("server called %d times, want 1 (cached)", callCount)
	}
}

func TestNewsItemJSON(t *testing.T) {
	item := NewsItem{Title: "Test headline", Link: "https://example.com", Source: "Test Feed"}
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
	if got.Source != item.Source {
		t.Errorf("Source = %q, want %q", got.Source, item.Source)
	}
}
