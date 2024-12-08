package updater

import (
	"time"

	"github.com/blang/semver/v4"
)

// UpdateChannel represents the release channel type
type UpdateChannel string

const (
	StableChannel UpdateChannel = "stable"
	BetaChannel   UpdateChannel = "beta"
)

// UpdateCache stores information about the last update check
type UpdateCache struct {
	LastCheck     time.Time     `json:"last_check"`
	LatestVersion string        `json:"latest_version"`
	Channel       UpdateChannel `json:"channel"`
	TTL           time.Duration `json:"ttl"`
}

// ReleaseInfo contains information about a release
type ReleaseInfo struct {
	Version      semver.Version `json:"version"`
	Channel      UpdateChannel  `json:"channel"`
	AssetURL     string        `json:"asset_url"`
	Signature    string        `json:"signature"`
	Checksum     string        `json:"checksum"`
	ReleaseNotes string        `json:"release_notes"`
}

// TelemetryData represents update telemetry information
type TelemetryData struct {
	Timestamp    time.Time     `json:"timestamp"`
	Success      bool          `json:"success"`
	FromVersion  string        `json:"from_version"`
	ToVersion    string        `json:"to_version"`
	Channel      UpdateChannel `json:"channel"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// RestartInfo contains information about the command to restart after update
type RestartInfo struct {
	Args        []string    `json:"args"`
	Environment []string    `json:"environment"`
	WorkingDir  string      `json:"working_dir"`
	Timestamp   time.Time   `json:"timestamp"`
}
