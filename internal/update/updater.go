package update

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Updater performs periodic update checks and applies them.
type Updater struct {
	CurrentVersion string
	GOOS           string
	GOARCH         string
	Client         *Client
	Replacer       func(binPath string, newBinary []byte) error
	Restarter      func(binPath string) error
}

// CheckAndUpdate checks for a newer version and applies it if found.
// Returns the new version string if updated, empty string if no update needed.
func (u *Updater) CheckAndUpdate() (string, error) {
	current, err := ParseVersion(u.CurrentVersion)
	if err != nil {
		log.Printf("update: skipping, current version %q is not a release", u.CurrentVersion)
		return "", nil
	}

	rel, err := u.Client.LatestRelease()
	if err != nil {
		return "", fmt.Errorf("checking for update: %w", err)
	}

	latest, err := ParseVersion(rel.TagName)
	if err != nil {
		return "", fmt.Errorf("parsing release tag %q: %w", rel.TagName, err)
	}

	if !latest.NewerThan(current) {
		log.Printf("update: %s is current", current)
		return "", nil
	}

	log.Printf("update: new version available %s -> %s", current, latest)

	assetName := fmt.Sprintf("resilver-%s-%s", u.GOOS, u.GOARCH)

	assetURL, err := rel.AssetURL(u.GOOS, u.GOARCH)
	if err != nil {
		return "", fmt.Errorf("finding asset: %w", err)
	}

	checksumURL, err := rel.ChecksumURL()
	if err != nil {
		return "", fmt.Errorf("finding checksums: %w", err)
	}

	checksums, err := u.Client.Download(checksumURL)
	if err != nil {
		return "", fmt.Errorf("downloading checksums: %w", err)
	}

	data, err := u.Client.Download(assetURL)
	if err != nil {
		return "", fmt.Errorf("downloading update: %w", err)
	}

	if err := VerifyChecksum(checksums, assetName, data); err != nil {
		return "", fmt.Errorf("integrity check failed: %w", err)
	}

	binPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locating binary: %w", err)
	}

	if err := u.Replacer(binPath, data); err != nil {
		return "", fmt.Errorf("replacing binary: %w", err)
	}

	log.Printf("update: binary replaced, restarting")
	if err := u.Restarter(binPath); err != nil {
		return "", fmt.Errorf("restarting: %w", err)
	}

	return latest.String(), nil
}

// Run starts the periodic update loop.
func (u *Updater) Run(interval time.Duration, stop <-chan struct{}) {
	u.check()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			u.check()
		case <-stop:
			return
		}
	}
}

func (u *Updater) check() {
	if _, err := u.CheckAndUpdate(); err != nil {
		log.Printf("update: %v", err)
	}
}

// ReplaceBinary atomically replaces the binary at binPath with newBinary.
// Keeps a backup at binPath+".bak" for rollback.
func ReplaceBinary(binPath string, newBinary []byte) error {
	info, err := os.Stat(binPath)
	if err != nil {
		return err
	}

	tmpPath := binPath + ".new"
	if err := os.WriteFile(tmpPath, newBinary, info.Mode().Perm()); err != nil {
		return err
	}

	bakPath := binPath + ".bak"
	if err := os.Rename(binPath, bakPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("backing up current binary: %w", err)
	}

	if err := os.Rename(tmpPath, binPath); err != nil {
		// Rollback: restore backup
		os.Rename(bakPath, binPath)
		return err
	}

	return nil
}
