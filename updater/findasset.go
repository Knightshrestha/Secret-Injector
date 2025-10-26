package updater

import (
	"fmt"
	"runtime"
)

func (u *Updater) findAsset(release *GitHubRelease, version string) (string, string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Expected filename: appname_version_os_arch.zip
	expectedName := fmt.Sprintf("%s_%s_%s_%s.zip", u.ExeName, version, goos, goarch)
	checksumName := fmt.Sprintf("%s_%s_checksums.txt", u.ExeName, version)

	var assetURL, checksumURL string

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			assetURL = asset.BrowserDownloadURL
		}
		if asset.Name == checksumName {
			checksumURL = asset.BrowserDownloadURL
		}
	}

	if assetURL == "" {
		return "", "", fmt.Errorf("no compatible asset found for %s/%s (looking for: %s)", goos, goarch, expectedName)
	}

	return assetURL, checksumURL, nil
}
