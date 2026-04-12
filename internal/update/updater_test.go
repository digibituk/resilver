package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const fakeBinary = "new-binary-content"

func fakeChecksum() string {
	h := sha256.Sum256([]byte(fakeBinary))
	return hex.EncodeToString(h[:])
}

func newTestServer(tag string) *httptest.Server {
	mux := http.NewServeMux()
	srv := httptest.NewUnstartedServer(mux)
	srv.Start()

	checksum := fakeChecksum()
	checksumBody := fmt.Sprintf("%s  resilver-linux-amd64\n", checksum)

	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"tag_name": %q,
			"assets": [
				{"name": "resilver-linux-amd64", "browser_download_url": "%s/download/resilver-linux-amd64"},
				{"name": "checksums.txt", "browser_download_url": "%s/download/checksums.txt"}
			]
		}`, tag, srv.URL, srv.URL)
	})
	mux.HandleFunc("/download/resilver-linux-amd64", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fakeBinary))
	})
	mux.HandleFunc("/download/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(checksumBody))
	})
	return srv
}

func TestCheckAndUpdateNewVersion(t *testing.T) {
	srv := newTestServer("v2.0.0")
	defer srv.Close()

	var replaceCalled bool
	var restartCalled bool
	var replacedBytes []byte

	u := &Updater{
		CurrentVersion: "v1.0.0",
		GOOS:           "linux",
		GOARCH:         "amd64",
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Owner: "owner", Repo: "repo"},
		Replacer: func(binPath string, data []byte) error {
			replaceCalled = true
			replacedBytes = data
			return nil
		},
		Restarter: func(binPath string) error {
			restartCalled = true
			return nil
		},
	}

	newVer, err := u.CheckAndUpdate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newVer != "v2.0.0" {
		t.Fatalf("got %q, want %q", newVer, "v2.0.0")
	}
	if !replaceCalled {
		t.Fatal("replacer was not called")
	}
	if !restartCalled {
		t.Fatal("restarter was not called")
	}
	if string(replacedBytes) != fakeBinary {
		t.Fatalf("got %q, want %q", replacedBytes, fakeBinary)
	}
}

func TestCheckAndUpdateAlreadyCurrent(t *testing.T) {
	srv := newTestServer("v1.0.0")
	defer srv.Close()

	var replaceCalled bool
	u := &Updater{
		CurrentVersion: "v1.0.0",
		GOOS:           "linux",
		GOARCH:         "amd64",
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Owner: "owner", Repo: "repo"},
		Replacer:       func(string, []byte) error { replaceCalled = true; return nil },
		Restarter:      func(string) error { return nil },
	}

	newVer, err := u.CheckAndUpdate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newVer != "" {
		t.Fatalf("got %q, want empty", newVer)
	}
	if replaceCalled {
		t.Fatal("replacer should not be called when already current")
	}
}

func TestCheckAndUpdateDevVersion(t *testing.T) {
	u := &Updater{
		CurrentVersion: "dev",
		GOOS:           "linux",
		GOARCH:         "amd64",
	}
	newVer, err := u.CheckAndUpdate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newVer != "" {
		t.Fatalf("got %q, want empty for dev version", newVer)
	}
}

func TestCheckAndUpdateDirtyVersion(t *testing.T) {
	u := &Updater{
		CurrentVersion: "v1.0.0-3-gabcdef",
		GOOS:           "linux",
		GOARCH:         "amd64",
	}
	newVer, err := u.CheckAndUpdate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newVer != "" {
		t.Fatalf("got %q, want empty for dirty version", newVer)
	}
}

func TestCheckAndUpdateGitHubError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	var replaceCalled bool
	u := &Updater{
		CurrentVersion: "v1.0.0",
		GOOS:           "linux",
		GOARCH:         "amd64",
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Owner: "owner", Repo: "repo"},
		Replacer:       func(string, []byte) error { replaceCalled = true; return nil },
		Restarter:      func(string) error { return nil },
	}

	_, err := u.CheckAndUpdate()
	if err == nil {
		t.Fatal("expected error for GitHub failure")
	}
	if replaceCalled {
		t.Fatal("replacer should not be called on error")
	}
}

func TestCheckAndUpdateNoMatchingAsset(t *testing.T) {
	srv := newTestServer("v2.0.0")
	defer srv.Close()

	u := &Updater{
		CurrentVersion: "v1.0.0",
		GOOS:           "darwin",
		GOARCH:         "arm64",
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Owner: "owner", Repo: "repo"},
		Replacer:       func(string, []byte) error { return nil },
		Restarter:      func(string) error { return nil },
	}

	_, err := u.CheckAndUpdate()
	if err == nil {
		t.Fatal("expected error for missing platform asset")
	}
}

func TestCheckAndUpdateChecksumMismatch(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"tag_name": "v2.0.0",
			"assets": [
				{"name": "resilver-linux-amd64", "browser_download_url": "%s/download/resilver-linux-amd64"},
				{"name": "checksums.txt", "browser_download_url": "%s/download/checksums.txt"}
			]
		}`, srv.URL, srv.URL)
	})
	mux.HandleFunc("/download/resilver-linux-amd64", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("tampered-binary"))
	})
	mux.HandleFunc("/download/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("%s  resilver-linux-amd64\n", fakeChecksum())))
	})

	var replaceCalled bool
	u := &Updater{
		CurrentVersion: "v1.0.0",
		GOOS:           "linux",
		GOARCH:         "amd64",
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Owner: "owner", Repo: "repo"},
		Replacer:       func(string, []byte) error { replaceCalled = true; return nil },
		Restarter:      func(string) error { return nil },
	}

	_, err := u.CheckAndUpdate()
	if err == nil {
		t.Fatal("expected error for checksum mismatch")
	}
	if replaceCalled {
		t.Fatal("replacer should not be called when checksum fails")
	}
}

func TestReplaceBinary(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "resilver")

	if err := os.WriteFile(binPath, []byte("old-content"), 0o755); err != nil {
		t.Fatal(err)
	}

	newContent := []byte("new-content")
	if err := ReplaceBinary(binPath, newContent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify contents replaced
	got, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-content" {
		t.Fatalf("got %q, want %q", got, "new-content")
	}

	// Verify permissions preserved
	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("got permissions %v, want 0755", info.Mode().Perm())
	}

	// Verify .new temp file does not persist
	if _, err := os.Stat(binPath + ".new"); !os.IsNotExist(err) {
		t.Fatal(".new temp file should not persist after successful replace")
	}

	// Verify backup exists
	bak, err := os.ReadFile(binPath + ".bak")
	if err != nil {
		t.Fatal("backup file should exist")
	}
	if string(bak) != "old-content" {
		t.Fatalf("backup got %q, want %q", bak, "old-content")
	}
}
