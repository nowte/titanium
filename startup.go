package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const titaniumVersion = "0.1.0-beta"

// githubRelease — GitHub API response shape
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// checkLatestVersion — fetches the latest release tag from GitHub.
// Returns ("", err) on failure, (tag, nil) on success.
func checkLatestVersion() (string, error) {
	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/nowte/titanium/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return strings.TrimSpace(release.TagName), nil
}

// UpdateAvailable — true if latest != current
func UpdateAvailable(latest string) bool {
	return latest != "" && latest != titaniumVersion
}

// buildStartupStatus — startup summary shown in chat on launch
func buildStartupStatus() string {
	var sb strings.Builder
	sb.WriteString("Titanium ready.\n")
	sb.WriteString("Type /help for available commands")
	return sb.String()
}
