package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/codingconcepts/grab/models"
)

func main() {
	c := &http.Client{
		Timeout: time.Second * 5,
	}

	release, err := getRelease(c, os.Args[1], os.Args[2])
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err = download(release); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func getRelease(c *http.Client, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=1", owner, repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating releases request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching releases: %w", err)
	}

	var releases []models.GitRelease
	if err = json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("reading releases: %w", err)
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	for _, asset := range releases[0].Assets {
		if strings.HasSuffix(asset.BrowserDownloadURL, "_"+runtime.GOOS) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no releases found for this os")
}

func download(url string) error {
	// Create file.
	file := path.Base(url)

	log.Printf("creating file %q", file)
	out, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", file, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("downloading release file %q: %w", url, err)
	}
	defer resp.Body.Close()

	// Write file.
	log.Printf("writing file %q", file)
	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file %q: %w", url, err)
	}

	// Make file executable.
	log.Printf("making file %q executable", file)
	cmd := exec.Command("chmod", "+x", file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("making file %q writable: %w", url, err)
	}

	return nil
}
