package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func (u *Updater) verifyChecksum(filePath, checksumURL string) error {
	// Download checksums file
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(checksumURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksums (status %d)", resp.StatusCode)
	}

	checksums, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Calculate file hash
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	fileHash := hex.EncodeToString(h.Sum(nil))

	// Find matching checksum
	exeName := u.ExeName
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}

	for line := range strings.SplitSeq(string(checksums), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.Contains(parts[1], exeName) {
			expectedHash := parts[0]
			if fileHash == expectedHash {
				return nil
			}
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, fileHash)
		}
	}

	return fmt.Errorf("checksum not found for %s", exeName)
}
