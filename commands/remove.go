package commands

import (
	"errors"
	"fmt"
	"log"

	"github.com/codingconcepts/grab/models"
	"github.com/codingconcepts/grab/state"
	"github.com/spf13/cobra"
)

// Remove an application
func Remove(cfg state.Config) func(cmd *cobra.Command, args []string) error {
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
