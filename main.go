package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/codingconcepts/grab/models"
	"github.com/codingconcepts/grab/state"
	"github.com/spf13/cobra"
)

const (
	stateFile = "grab_state.json"
	grabBin   = "bin"

	releasesPerPage = 100
)

func main() {
	log.SetFlags(0)

	dir := "./grab"
	config, err := ensureGrabDir(dir)
	if err != nil {
		log.Fatalf("error ensuring grab directory: %v", err)
	}

	c := &http.Client{
		Timeout: time.Second * 5,
	}

	installCmd := &cobra.Command{
		Use:     "install",
		Short:   "Installs a package",
		Example: "grab install codingconcepts pa55 [VERSION]",
		Args:    cobra.RangeArgs(2, 3),
		RunE:    install(c, config),
	}

	updateCmd := &cobra.Command{
		Use:     "update",
		Short:   "Updates a package",
		Example: "grab update codingconcepts pa55 [VERSION]",
		Args:    cobra.RangeArgs(2, 3),
		RunE:    update(c, config),
	}

	removeCmd := &cobra.Command{
		Use:     "remove",
		Short:   "Removes a package",
		Example: "grab remove codingconcepts pa55",
		Args:    cobra.ExactArgs(2),
		RunE:    remove(config),
	}

	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(installCmd, updateCmd, removeCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func install(c *http.Client, cfg state.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		repo := args[1]

		var version string
		if len(args) > 2 {
			version = args[2]
		}

		var release models.Release
		var err error

		// Get the required version of the release (or latest if not specified).
		if version != "" {
			release, err = getVersionedRelease(c, owner, repo, version)
			if err != nil {
				if errors.As(err, &models.ErrNotFound{}) {
					log.Println(err)
					return nil
				}
				return fmt.Errorf("getting release: %w", err)
			}
		} else {
			release, err = getLatestRelease(c, owner, repo)
			if err != nil {
				return fmt.Errorf("getting release: %w", err)
			}
		}

		release.Owner = owner
		release.Repo = repo
		release.InstalledPath = path.Join(cfg.BinDirPath, path.Base(release.URL))

		// Check whether a version of this file has already been downloaded and
		// if so, bail with a message telling the user to update instead.
		if _, err := os.Stat(release.InstalledPath); !os.IsNotExist(err) {
			log.Println("a version for this app already exists, consider running update instead")
			return nil
		}

		// Download the version found.
		if err = download(release, cfg); err != nil {
			return fmt.Errorf("downloading release: %w", err)
		}

		// Update state file.
		if err = state.WriteStateFile(cfg.StateFilePath, release); err != nil {
			return fmt.Errorf("writing state file: %w", err)
		}

		return nil
	}
}

func update(c *http.Client, cfg state.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		repo := args[1]

		current, err := state.GetRelease(cfg.StateFilePath, owner, repo)
		if err != nil {
			if errors.As(err, &models.ErrNotFound{}) {
				log.Printf("package not installed, consider using install instead")
				return nil
			}
			return fmt.Errorf("getting current version: %w", err)
		}

		var version string
		if len(args) > 2 {
			version = args[2]
		}

		var release models.Release

		// Get the required version of the release (or latest if not specified).
		if version != "" {
			// If we've already got this version, bail now.
			if version == current.Version {
				log.Printf("version %s is already installed, call update without a version to install latest", version)
				return nil
			}

			release, err = getVersionedRelease(c, owner, repo, version)
			if err != nil {
				if errors.As(err, &models.ErrNotFound{}) {
					log.Println(err)
					return nil
				}
				return fmt.Errorf("getting release: %w", err)
			}
		} else {
			release, err = getLatestRelease(c, owner, repo)
			if err != nil {
				return fmt.Errorf("getting release: %w", err)
			}
		}
		release.Owner = owner
		release.Repo = repo
		release.InstalledPath = path.Join(cfg.BinDirPath, path.Base(release.URL))

		// Download the version found.
		if err = download(release, cfg); err != nil {
			return fmt.Errorf("downloading release: %w", err)
		}

		// Update state file.
		if err = state.WriteStateFile(cfg.StateFilePath, release); err != nil {
			return fmt.Errorf("writing state file: %w", err)
		}

		return nil
	}
}

func remove(cfg state.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		repo := args[1]

		// Check for an existing package.
		release, err := state.GetRelease(cfg.StateFilePath, owner, repo)
		if err != nil {
			if errors.As(err, &models.ErrNotFound{}) {
				log.Printf("no installed packages for owner=%q repo=%q", owner, repo)
				return nil
			}
		}

		if err = removeRelease(cfg, owner, repo, release); err != nil {
			return fmt.Errorf("removing release: %w", err)
		}

		return nil
	}
}

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

func getVersionedRelease(c *http.Client, owner, repo, version string) (models.Release, error) {
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

func ensureGrabDir(baseDir string) (state.Config, error) {
	// Initialise the config parameters.
	config := state.Config{
		StateFilePath: path.Join(baseDir, stateFile),
		BinDirPath:    path.Join(baseDir, grabBin),
	}

	// If the grab dir already exists, there's nothing to do.
	if _, err := os.Stat(baseDir); !os.IsNotExist(err) {
		return config, nil
	}

	// Create the grab dir.
	if err := os.MkdirAll(config.BinDirPath, os.ModePerm); err != nil {
		return config, fmt.Errorf("creating directory %q: %w", baseDir, err)
	}

	// Create the state file.
	fullPath := path.Join(baseDir, stateFile)
	if err := ioutil.WriteFile(fullPath, []byte(`{}`), 0644); err != nil {
		return config, fmt.Errorf("writing state file %q: %w", fullPath, err)
	}

	return config, nil
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
