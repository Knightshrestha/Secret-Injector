package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Update performs the self-update
func (u *Updater) Update() error {
	fmt.Println("Checking for updates...")

	hasUpdate, latestVer, err := u.CheckForUpdate()
	if err != nil {
		return err
	}

	if !hasUpdate {
		fmt.Printf("Already up to date (v%s)\n", u.CurrentVer)
		return nil
	}

	fmt.Printf("New version available: v%s (current: v%s)\n", latestVer, u.CurrentVer)
	fmt.Println("Downloading update...")

	release, err := u.fetchLatestRelease()
	if err != nil {
		return err
	}

	// Find the appropriate asset for this platform
	assetURL, checksumURL, err := u.findAsset(release, latestVer)
	if err != nil {
		return err
	}

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	tmpPath := filepath.Join(exeDir, u.ExeName+".tmp")
	oldPath := filepath.Join(exeDir, u.ExeName+".old")

	// Download to temporary file
	if err := u.downloadAndExtract(assetURL, tmpPath); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tmpPath) // Clean up tmp file on error

	// Verify checksum if available
	if checksumURL != "" {
		fmt.Println("Verifying checksum...")
		if err := u.verifyChecksum(tmpPath, checksumURL); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		fmt.Println("Checksum verified successfully")
	}

	// Rename old executable
	if _, err := os.Stat(oldPath); err == nil {
		if err := os.Remove(oldPath); err != nil {
			return fmt.Errorf("failed to remove old backup: %w", err)
		}
	}

	if err := os.Rename(exePath, oldPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %w", err)
	}

	// Rename new executable
	if err := os.Rename(tmpPath, exePath); err != nil {
		// Try to restore old executable
		os.Rename(oldPath, exePath)
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(exePath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}

	fmt.Printf("Successfully updated to v%s\n", latestVer)
	fmt.Println("Old version backed up to:", filepath.Base(oldPath))
	return nil
}
