package commands

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/codingconcepts/grab/models"
	"github.com/codingconcepts/grab/state"
	"github.com/spf13/cobra"
)

// Install an application
func Install(c *http.Client, cfg state.Config, releasesPerPage int) func(cmd *cobra.Command, args []string) error {
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
			release, err = getVersionedRelease(c, owner, repo, version, releasesPerPage)
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
