package commands

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

	"github.com/codingconcepts/grab/models"
	"github.com/codingconcepts/grab/state"
)

func getLatestRelease(c *http.Client, owner, repo string) (models.Release, error) {
	releases, err := getReleases(c, owner, repo, 1, 1)
	if err != nil {
		return models.Release{}, fmt.Errorf("getting first release: %w", err)
	}

	if len(releases) == 0 {
		return models.Release{}, fmt.Errorf("no releases found")
	}

	for _, asset := range releases[0].Assets {
		if strings.HasSuffix(asset.BrowserDownloadURL, "_"+runtime.GOOS) {
			return models.Release{
				Version: releases[0].TagName,
				URL:     asset.BrowserDownloadURL,
			}, nil
		}
	}

	return models.Release{}, fmt.Errorf("no releases found for %s", runtime.GOOS)
}

func getVersionedRelease(c *http.Client, owner, repo, version string, releasesPerPage int) (models.Release, error) {
	page := 1
	for {
		var pageReleases []models.GitRelease
		pageReleases, err := getReleases(c, owner, repo, page, releasesPerPage)
		if err != nil {
			return models.Release{}, fmt.Errorf("getting release page %d: %w", page, err)
		}

		// If this is the first run of the loop, there aren't any releases at all,
		// otherwise, we've reached the end and not found any releases with the
		// required version.
		if len(pageReleases) == 0 {
			return models.Release{}, models.MakeErrNotFound("no releases for version %s", version)
		}

		for _, release := range pageReleases {
			if release.TagName == version {
				for _, asset := range release.Assets {
					if strings.HasSuffix(asset.BrowserDownloadURL, "_"+runtime.GOOS) {
						return models.Release{
							Version: release.TagName,
							URL:     asset.BrowserDownloadURL,
						}, nil
					}
				}

				return models.Release{}, fmt.Errorf("no %s installation available", runtime.GOOS)
			}
		}

		page++
	}
}

func getReleases(c *http.Client, owner, repo string, page, perPage int) ([]models.GitRelease, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases?per_page=%d&page=%d",
		owner, repo, perPage, page)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating releases request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching releases: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if until, ok := models.ParseRateLimitResetTime(resp.Header); ok {
			return nil, fmt.Errorf("rate-limit exceeded, try again in %s", until)
		}
	}

	var releases []models.GitRelease
	if err = json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("reading releases: %w", err)
	}

	return releases, nil
}

func download(release models.Release, cfg state.Config) error {
	// Download file.
	resp, err := http.Get(release.URL)
	if err != nil {
		return fmt.Errorf("downloading release file %q: %w", release.URL, err)
	}
	defer resp.Body.Close()

	// Create file.
	fileName := path.Join(cfg.BinDirPath, path.Base(release.URL))

	log.Printf("creating file %q", fileName)
	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", fileName, err)
	}
	defer out.Close()

	// Write file.
	log.Printf("writing file %q", fileName)
	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file %q: %w", fileName, err)
	}

	// Make file executable.
	log.Printf("making file %q executable", fileName)
	cmd := exec.Command("chmod", "+x", fileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("making file %q writable: %w", release.URL, err)
	}

	return nil
}

func removeRelease(cfg state.Config, owner, repo string, release models.Release) error {
	log.Printf("removing file %q", release.InstalledPath)
	if err := os.Remove(release.InstalledPath); err != nil {
		return fmt.Errorf("removing file %q: %w", release.InstalledPath, err)
	}

	if err := state.RemoveRelease(cfg.StateFilePath, release); err != nil {
		return fmt.Errorf("removing from state file: %w", err)
	}

	return nil
}
