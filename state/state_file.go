package state

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codingconcepts/grab/models"
)

// Config contains the parameters we'll need for grab's commands.
type Config struct {
	StateFilePath string
	BinDirPath    string
}

// Owners is a map of owner names to Repos.
type Owners map[string]Repos

// Repos is a map of repo names to versions.
type Repos map[string]models.Release

// CreateStateFile creates the state file that will be used to store
// application versions.
func CreateStateFile(path string) error {
	return nil
}

// ListOwners reads the state file and returns its contents.
func ListOwners(path string) (Owners, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening state file %q: %w", path, err)
	}
	defer f.Close()

	var owners Owners
	if err = json.NewDecoder(f).Decode(&owners); err != nil {
		return nil, fmt.Errorf("reading state file: %q: %w", path, err)
	}

	return owners, nil
}

// GetRelease returns a release for an owner/repo pair or an ErrNotFound if
// one is not found.
func GetRelease(path, owner, repo string) (models.Release, error) {
	owners, err := ListOwners(path)
	if err != nil {
		return models.Release{}, fmt.Errorf("listing installed packages: %w", err)
	}

	if _, ok := owners[owner]; !ok {
		return models.Release{}, models.NewErrNotFound("no packages for owner %q", owner)
	}

	release, ok := owners[owner][repo]
	if !ok {
		return models.Release{}, models.NewErrNotFound("no packages for repo %q", repo)
	}

	return release, nil
}

// WriteStateFile writes any changes to the state file.
func WriteStateFile(path string, release models.Release) error {
	// Read the existing state file.
	owners, err := ListOwners(path)
	if err != nil {
		return fmt.Errorf("reading state file %q: %w", path, err)
	}

	// Add the new owner/repo.
	if _, ok := owners[release.Owner]; !ok {
		owners[release.Owner] = Repos{}
	}
	owners[release.Owner][release.Repo] = release

	// Overwrite the file with the updated content.
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening state file %q: %w", path, err)
	}
	defer f.Close()

	contents, err := json.MarshalIndent(owners, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling json: %w", err)
	}

	if _, err = f.Write(contents); err != nil {
		return fmt.Errorf("writing state file: %q: %w", path, err)
	}

	return nil
}
