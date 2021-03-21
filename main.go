package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/codingconcepts/grab/commands"
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
		RunE:    commands.Install(c, config, releasesPerPage),
	}

	updateCmd := &cobra.Command{
		Use:     "update",
		Short:   "Updates a package",
		Example: "grab update codingconcepts pa55 [VERSION]",
		Args:    cobra.RangeArgs(2, 3),
		RunE:    commands.Update(c, config, releasesPerPage),
	}

	removeCmd := &cobra.Command{
		Use:     "remove",
		Short:   "Removes a package",
		Example: "grab remove codingconcepts pa55",
		Args:    cobra.ExactArgs(2),
		RunE:    commands.Remove(config),
	}

	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(installCmd, updateCmd, removeCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
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
