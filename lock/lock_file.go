package lock

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codingconcepts/grab/models"
)

// Config contains the parameters we'll need for grab's commands.
type Config struct {
	LockFilePath string
	BinDirPath   string
}

// Owners is a map of owner names to Repos.
type Owners map[string]Repos

// Repos is a map of repo names to versions.
type Repos map[string]string

// CreateLockFile creates the lock file that will be used to store
// application versions.
func CreateLockFile(path string) error {
	return nil
}

// ReadLockFile reads the lock file and returns its contents.
func ReadLockFile(path string) (Owners, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening lock file %q: %w", path, err)
	}
	defer f.Close()

	var owners Owners
	if err = json.NewDecoder(f).Decode(&owners); err != nil {
		return nil, fmt.Errorf("reading lock file: %q: %w", path, err)
	}

	return owners, nil
}

// WriteLockFile writes any changes to the lock file.
func WriteLockFile(path string, release models.Release) error {
	// Read the existing lock file.
	owners, err := ReadLockFile(path)
	if err != nil {
		return fmt.Errorf("reading lock file %q: %w", path, err)
	}

	// Add the new owner/repo.
	if _, ok := owners[release.Owner]; !ok {
		owners[release.Owner] = Repos{}
	}
	owners[release.Owner][release.Repo] = release.Version

	// Overwrite the file with the updated content.
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening lock file %q: %w", path, err)
	}
	defer f.Close()

	if err = json.NewEncoder(f).Encode(owners); err != nil {
		return fmt.Errorf("writing lock file: %q: %w", path, err)
	}

	return nil
}
