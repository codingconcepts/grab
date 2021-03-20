package models

// Release is our internal representation of a release, containing just
// the fields we're interesting in.
type Release struct {
	Owner   string
	Repo    string
	Version string
	URL     string
}
