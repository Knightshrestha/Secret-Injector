package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func (u *Updater) downloadAndExtract(url, destPath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Download to temporary zip file
	tmpZip := destPath + ".zip"
	defer os.Remove(tmpZip)

	out, err := os.Create(tmpZip)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return err
	}

	// Extract executable from zip
	return u.extractExe(tmpZip, destPath)
}
