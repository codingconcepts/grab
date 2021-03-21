package commands

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/codingconcepts/grab/models"
	"github.com/codingconcepts/grab/state"
	"github.com/spf13/cobra"
)

// Update an application.
func Update(c *http.Client, cfg state.Config, releasesPerPage int) func(cmd *cobra.Command, args []string) error {
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
