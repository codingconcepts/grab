package models

// Release is our internal representation of a release, containing just
// the fields we're interesting in.
type Release struct {
	Owner         string `json:"owner"`
	Repo          string `json:"repo"`
	Version       string `json:"version"`
	URL           string `json:"url"`
	InstalledPath string `json:"installed_path"`
}
