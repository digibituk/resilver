package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/latest" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tag_name": "v1.2.3",
			"assets": [
				{"name": "resilver-linux-arm64", "browser_download_url": "https://example.com/resilver-linux-arm64"},
				{"name": "resilver-linux-amd64", "browser_download_url": "https://example.com/resilver-linux-amd64"}
			]
		}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Owner: "owner", Repo: "repo"}
	rel, err := c.LatestRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel.TagName != "v1.2.3" {
		t.Fatalf("got tag %q, want %q", rel.TagName, "v1.2.3")
	}
	if len(rel.Assets) != 2 {
		t.Fatalf("got %d assets, want 2", len(rel.Assets))
	}
}

func TestLatestReleaseHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Owner: "owner", Repo: "repo"}
	_, err := c.LatestRelease()
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestAssetURL(t *testing.T) {
	rel := Release{
		Assets: []Asset{
			{Name: "resilver-linux-arm64", BrowserDownloadURL: "https://example.com/resilver-linux-arm64"},
			{Name: "resilver-linux-amd64", BrowserDownloadURL: "https://example.com/resilver-linux-amd64"},
		},
	}
	url, err := rel.AssetURL("linux", "arm64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://example.com/resilver-linux-arm64" {
		t.Fatalf("got %q, want arm64 URL", url)
	}
}

func TestAssetURLNotFound(t *testing.T) {
	rel := Release{
		Assets: []Asset{
			{Name: "resilver-linux-amd64", BrowserDownloadURL: "https://example.com/resilver-linux-amd64"},
		},
	}
	_, err := rel.AssetURL("darwin", "arm64")
	if err == nil {
		t.Fatal("expected error for missing asset")
	}
}

func TestDownload(t *testing.T) {
	content := []byte("fake-binary-content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTPClient: srv.Client()}
	data, err := c.Download(srv.URL + "/resilver-linux-arm64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("got %q, want %q", data, content)
	}
}

func TestChecksumURL(t *testing.T) {
	rel := Release{
		Assets: []Asset{
			{Name: "resilver-linux-arm64", BrowserDownloadURL: "https://example.com/resilver-linux-arm64"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
		},
	}
	url, err := rel.ChecksumURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://example.com/checksums.txt" {
		t.Fatalf("got %q", url)
	}
}

func TestChecksumURLNotFound(t *testing.T) {
	rel := Release{
		Assets: []Asset{
			{Name: "resilver-linux-arm64", BrowserDownloadURL: "https://example.com/resilver-linux-arm64"},
		},
	}
	_, err := rel.ChecksumURL()
	if err == nil {
		t.Fatal("expected error for missing checksums.txt")
	}
}

func TestVerifyChecksumValid(t *testing.T) {
	data := []byte("binary-content")
	hash := sha256.Sum256(data)
	checksums := fmt.Sprintf("%s  resilver-linux-arm64\n", hex.EncodeToString(hash[:]))

	if err := VerifyChecksum([]byte(checksums), "resilver-linux-arm64", data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyChecksumMismatch(t *testing.T) {
	checksums := "0000000000000000000000000000000000000000000000000000000000000000  resilver-linux-arm64\n"
	err := VerifyChecksum([]byte(checksums), "resilver-linux-arm64", []byte("tampered"))
	if err == nil {
		t.Fatal("expected error for checksum mismatch")
	}
}

func TestVerifyChecksumMissingEntry(t *testing.T) {
	checksums := "abcdef1234567890  resilver-linux-amd64\n"
	err := VerifyChecksum([]byte(checksums), "resilver-linux-arm64", []byte("data"))
	if err == nil {
		t.Fatal("expected error for missing checksum entry")
	}
}
