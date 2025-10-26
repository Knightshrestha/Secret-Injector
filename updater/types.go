package updater

import (
	"time"
)

const (
	githubAPIURL = "https://api.github.com/repos/%s/%s/releases/latest"
	timeout      = 30 * time.Second
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type Updater struct {
	Owner      string // GitHub username or org
	Repo       string // Repository name
	CurrentVer string // Current version (e.g., "0.0.1")
	ExeName    string // Executable name without extension
}
