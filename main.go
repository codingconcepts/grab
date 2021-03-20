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
	"github.com/spf13/cobra"
)

func main() {
	log.SetFlags(0)

	c := &http.Client{
		Timeout: time.Second * 5,
	}

	installCmd := &cobra.Command{
		Use:     "install",
		Short:   "Installs a package",
		Example: "grab install codingconcepts pa55 [VERSION]",
		Args:    cobra.RangeArgs(2, 3),
		RunE:    install(c),
	}

	updateCmd := &cobra.Command{
		Use:     "update",
		Short:   "Updates a package",
		Example: "grab update codingconcepts pa55 [VERSION]",
		Args:    cobra.RangeArgs(2, 3),
		RunE:    update(c),
	}

	deleteCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Deletes a package",
		Example: "grab delete codingconcepts pa55",
		Args:    cobra.ExactArgs(2),
		RunE:    delete,
	}

	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(installCmd, updateCmd, deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func install(c *http.Client) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		repo := args[1]

		var version string
		if len(args) > 2 {
			version = args[2]
		}

		release, err := getRelease(c, owner, repo, version)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if err = download(release); err != nil {
			log.Fatalf("error: %v", err)
		}
		return nil
	}
}

func update(c *http.Client) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func delete(cmd *cobra.Command, args []string) error {
	return nil
}

func getRelease(c *http.Client, owner, repo, version string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=100&page=%d", owner, repo, 1)

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
