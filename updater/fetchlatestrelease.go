package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (u *Updater) fetchLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf(githubAPIURL, u.Owner, u.Repo)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}
