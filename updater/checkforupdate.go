package updater

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"strings"
)

// CheckForUpdate checks if a newer version is available
func (u *Updater) CheckForUpdate() (bool, string, error) {
	release, err := u.fetchLatestRelease()
	if err != nil {
		return false, "", fmt.Errorf("failed to fetch latest release: %w", err)
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")

	current, err := version.NewVersion(u.CurrentVer)
	if err != nil {
		return false, "", fmt.Errorf("invalid current version: %w", err)
	}

	latest, err := version.NewVersion(latestVer)
	if err != nil {
		return false, "", fmt.Errorf("invalid latest version: %w", err)
	}

	if latest.GreaterThan(current) {
		return true, latestVer, nil
	}

	return false, latestVer, nil
}
