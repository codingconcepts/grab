package main

import (
	"encoding/json"
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

	"github.com/codingconcepts/grab/lock"
	"github.com/codingconcepts/grab/models"
	"github.com/spf13/cobra"
)

const (
	lockFile = "grab_lock"
	grabBin  = "bin"
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

	deleteCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Deletes a package",
		Example: "grab delete codingconcepts pa55",
		Args:    cobra.ExactArgs(2),
		RunE:    delete(config),
	}

	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(installCmd, updateCmd, deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func install(c *http.Client, cfg lock.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		repo := args[1]

		var version string
		if len(args) > 2 {
			version = args[2]
		}

		// Get the required version of the release (or latest if not specified).
		release, err := getRelease(c, owner, repo, version)
		if err != nil {
			return fmt.Errorf("getting release: %w", err)
		}
		release.Owner = owner
		release.Repo = repo

		// Check whether a version of this file has already been downloaded and
		// if so, bail with a message telling the user to update instead.
		fullPath := path.Join(cfg.BinDirPath, path.Base(release.URL))
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			log.Println("a version for this app already exists, consider running update instead")
			return nil
		}

		// Download the version found.
		if err = download(release, cfg); err != nil {
			return fmt.Errorf("downloading release: %w", err)
		}

		// Update lock file.
		if err = lock.WriteLockFile(cfg.LockFilePath, release); err != nil {
			return fmt.Errorf("writing locking file: %w", err)
		}

		return nil
	}
}

func update(c *http.Client, cfg lock.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func delete(cfg lock.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func getRelease(c *http.Client, owner, repo, version string) (models.Release, error) {
	perPage := 1
	if version != "" {
		perPage = 100
	}

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases?per_page=%d&page=%d",
		owner, repo, perPage, 1)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return models.Release{}, fmt.Errorf("creating releases request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return models.Release{}, fmt.Errorf("fetching releases: %w", err)
	}

	// TODO: Use if we've received a version parameter and it's not on the page.
	// log.Println(linkheader.Parse(resp.Header.Get("Link")))

	var releases []models.GitRelease
	if err = json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return models.Release{}, fmt.Errorf("reading releases: %w", err)
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

	return models.Release{}, fmt.Errorf("no releases found for this os")
}

func ensureGrabDir(baseDir string) (lock.Config, error) {
	// Initialise the config parameters.
	config := lock.Config{
		LockFilePath: path.Join(baseDir, lockFile),
		BinDirPath:   path.Join(baseDir, grabBin),
	}

	// If the grab dir already exists, there's nothing to do.
	if _, err := os.Stat(baseDir); !os.IsNotExist(err) {
		return config, nil
	}

	// Create the grab dir.
	if err := os.MkdirAll(config.BinDirPath, os.ModePerm); err != nil {
		return config, fmt.Errorf("creating directory %q: %w", baseDir, err)
	}

	// Create the lock file.
	fullPath := path.Join(baseDir, lockFile)
	if err := ioutil.WriteFile(fullPath, []byte(`{}`), 0644); err != nil {
		return config, fmt.Errorf("writing lock file %q: %w", fullPath, err)
	}

	return config, nil
}

func download(release models.Release, cfg lock.Config) error {
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
