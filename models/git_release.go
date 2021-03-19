package models

import "time"

// GitRelease describes a particular version of a release.
type GitRelease struct {
	TagName     string     `json:"tag_name"`
	Name        string     `json:"name"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt time.Time  `json:"published_at"`
	Assets      []GitAsset `json:"assets"`
}

// GitAsset describes the file in a particular version of a release.
type GitAsset struct {
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}
