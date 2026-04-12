package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Release represents a GitHub Release.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a downloadable file in a GitHub Release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Client checks GitHub for new releases and downloads binaries.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Owner      string
	Repo       string
}

// LatestRelease fetches the latest release from GitHub.
func (c *Client) LatestRelease() (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.BaseURL, c.Owner, c.Repo)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return Release{}, fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return Release{}, fmt.Errorf("decoding release: %w", err)
	}
	return rel, nil
}

// AssetURL returns the download URL for the binary matching the given OS and architecture.
func (r Release) AssetURL(goos, goarch string) (string, error) {
	name := fmt.Sprintf("resilver-%s-%s", goos, goarch)
	for _, a := range r.Assets {
		if a.Name == name {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no asset found for %s/%s", goos, goarch)
}

// MaxDownloadSize is the maximum allowed binary size (16 MB).
const MaxDownloadSize = 16 << 20

// Download fetches the binary from the given URL.
func (c *Client) Download(url string) ([]byte, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, MaxDownloadSize))
}

// ChecksumURL returns the download URL for the checksums file.
func (r Release) ChecksumURL() (string, error) {
	for _, a := range r.Assets {
		if a.Name == "checksums.txt" {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no checksums.txt asset found")
}

// VerifyChecksum checks that the SHA-256 hash of data matches the expected
// hash for the given asset name in the checksums file content.
// Checksums file format: "<hex-hash>  <filename>\n" per line.
func VerifyChecksum(checksums []byte, assetName string, data []byte) error {
	hash := sha256.Sum256(data)
	got := hex.EncodeToString(hash[:])

	for _, line := range strings.Split(string(checksums), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == assetName {
			if parts[0] != got {
				return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", assetName, parts[0], got)
			}
			return nil
		}
	}
	return fmt.Errorf("no checksum entry for %s", assetName)
}
