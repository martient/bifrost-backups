package updater

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v35/github"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/net/context"
)

// Updater handles the auto-update functionality
type Updater struct {
	CurrentVersion string
	GithubRepo     string
	CacheFile      string
	PublicKey      ed25519.PublicKey
	Channel        UpdateChannel
	TelemetryPath  string
	RestartPath    string
	cache          *UpdateCache
	client         *github.Client
}

// New creates a new Updater instance
func New(currentVersion, githubRepo string, channel UpdateChannel) (*Updater, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Join(configDir, "bifrost-backup")
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return nil, err
	}

	return &Updater{
		CurrentVersion: currentVersion,
		GithubRepo:     githubRepo,
		CacheFile:      filepath.Join(baseDir, "update-cache.json"),
		Channel:        channel,
		TelemetryPath:  filepath.Join(baseDir, "update-telemetry.json"),
		RestartPath:    filepath.Join(baseDir, "restart-info.json"),
		client:         github.NewClient(nil),
	}, nil
}

// Check looks for available updates
func (u *Updater) Check() (*ReleaseInfo, error) {
	// Check if we need to update based on the cache
	if !u.needsCheck() {
		return nil, nil
	}

	// Get the latest release
	release, err := u.getLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	if err := u.updateCache(release); err != nil {
		return nil, fmt.Errorf("failed to update cache: %w", err)
	}

	// Compare versions
	currentVersion, err := semver.Parse(u.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version: %w", err)
	}

	if release.Version.GT(currentVersion) {
		return release, nil
	}

	return nil, nil
}

// Update performs the actual update
func (u *Updater) Update(release *ReleaseInfo) error {
	// Start telemetry
	telemetry := &TelemetryData{
		Timestamp:   time.Now(),
		FromVersion: u.CurrentVersion,
		ToVersion:   release.Version.String(),
		Channel:     u.Channel,
		Success:     false,
	}
	defer func() {
		if err := u.saveTelemetry(telemetry); err != nil {
			log.Printf("Failed to save telemetry: %v", err)
		}
	}()

	// Get the executable path
	exe, err := os.Executable()
	if err != nil {
		telemetry.ErrorMessage = err.Error()
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a temporary file for the download
	tmpFile, err := os.CreateTemp("", "bifrost-backup-*")
	if err != nil {
		telemetry.ErrorMessage = err.Error()
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Download the binary
	if err := u.downloadBinary(release, tmpFile); err != nil {
		telemetry.ErrorMessage = err.Error()
		if rollbackErr := u.rollback(exe); rollbackErr != nil {
			log.Printf("Failed to rollback: %v", rollbackErr)
		}
		return fmt.Errorf("failed to download binary: %w", err)
	}

	// Read the downloaded binary for verification
	binary, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		telemetry.ErrorMessage = err.Error()
		if rollbackErr := u.rollback(exe); rollbackErr != nil {
			log.Printf("Failed to rollback: %v", rollbackErr)
		}
		return fmt.Errorf("failed to read downloaded binary: %w", err)
	}

	// Verify the checksum
	if err := u.verifyChecksum(binary, release.Checksum); err != nil {
		telemetry.ErrorMessage = err.Error()
		if rollbackErr := u.rollback(exe); rollbackErr != nil {
			log.Printf("Failed to rollback: %v", rollbackErr)
		}
		return fmt.Errorf("failed to verify checksum: %w", err)
	}

	// Verify the signature
	if err := u.verifySignature(binary, release.Signature); err != nil {
		telemetry.ErrorMessage = err.Error()
		if rollbackErr := u.rollback(exe); rollbackErr != nil {
			log.Printf("Failed to rollback: %v", rollbackErr)
		}
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	// Replace the old binary with the new one
	if err := u.replaceBinary(tmpFile, exe); err != nil {
		telemetry.ErrorMessage = err.Error()
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	telemetry.Success = true
	return nil
}

// GetChangelog retrieves the changelog for the release
func (u *Updater) GetChangelog(release *ReleaseInfo) string {
	return release.ReleaseNotes
}

func (u *Updater) rollback(exe string) error {
	backupPath := exe + ".backup"
	if _, err := os.Stat(backupPath); err == nil {
		return os.Rename(backupPath, exe)
	}
	return fmt.Errorf("no backup found")
}

func (u *Updater) verifyChecksum(binary []byte, expectedChecksum string) error {
	hash := sha256.Sum256(binary)
	actualChecksum := hex.EncodeToString(hash[:])
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum verification failed")
	}
	return nil
}

func (u *Updater) verifySignature(binary []byte, signature string) error {
	if u.PublicKey == nil {
		return fmt.Errorf("no public key available")
	}
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}
	if !ed25519.Verify(u.PublicKey, binary, sigBytes) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

func (u *Updater) downloadBinary(release *ReleaseInfo, tmpFile *os.File) error {
	resp, err := http.Get(release.AssetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	return err
}

func (u *Updater) replaceBinary(tmpFile *os.File, exe string) error {
	tmpFile.Close()
	return os.Rename(tmpFile.Name(), exe)
}

func (u *Updater) saveTelemetry(data *TelemetryData) error {
	var telemetry []TelemetryData

	// Read existing telemetry
	if content, err := os.ReadFile(u.getTelemetryPath()); err == nil {
		if err := json.Unmarshal(content, &telemetry); err != nil {
			return fmt.Errorf("failed to unmarshal telemetry: %w", err)
		}
	}

	// Append new telemetry
	telemetry = append(telemetry, *data)

	// Write back to file
	content, err := json.Marshal(telemetry)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	return os.WriteFile(u.getTelemetryPath(), content, 0640)
}

// func (u *Updater) loadTelemetry() (*TelemetryData, error) {
// 	content, err := os.ReadFile(u.getTelemetryPath())
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return &TelemetryData{}, nil
// 		}
// 		return nil, err
// 	}

// 	var telemetry TelemetryData
// 	if err := json.Unmarshal(content, &telemetry); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal telemetry: %w", err)
// 	}

// 	return &telemetry, nil
// }

func (u *Updater) needsCheck() bool {
	return time.Since(u.cache.LastCheck) > u.cache.TTL
}

func (u *Updater) getLatestRelease() (*ReleaseInfo, error) {
	owner, repo, found := strings.Cut(u.GithubRepo, "/")
	if !found {
		return nil, fmt.Errorf("invalid github repo format: %s", u.GithubRepo)
	}

	// List all releases
	releases, _, err := u.client.Repositories.ListReleases(context.Background(), owner, repo, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	// Filter releases based on channel
	var latestRelease *github.RepositoryRelease
	for _, release := range releases {
		// Skip draft releases
		if release.GetDraft() {
			continue
		}

		// For beta channel, include pre-releases
		// For stable channel, exclude pre-releases
		if u.Channel == BetaChannel {
			if latestRelease == nil || release.GetPrerelease() {
				latestRelease = release
			}
		} else if !release.GetPrerelease() {
			latestRelease = release
			break
		}
	}

	if latestRelease == nil {
		return nil, nil
	}

	return u.parseReleaseInfo(latestRelease)
}

func (u *Updater) parseReleaseInfo(release *github.RepositoryRelease) (*ReleaseInfo, error) {
	version, err := semver.ParseTolerant(release.GetTagName())
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %w", err)
	}

	// Determine platform-specific asset name
	assetNamePrefix := fmt.Sprintf("bifrost-backups_%s_%s", runtime.GOOS, getPlatformArch())
	var assetURL string
	var checksumAsset *github.ReleaseAsset

	for _, asset := range release.Assets {
		name := asset.GetName()
		if strings.HasPrefix(name, assetNamePrefix) {
			assetURL = asset.GetBrowserDownloadURL()
		} else if name == "bifrost-backups_checksums.txt" {
			checksumAsset = asset
		}
	}

	if assetURL == "" {
		return nil, fmt.Errorf("no compatible binary found for %s_%s", runtime.GOOS, getPlatformArch())
	}

	// Get checksum
	checksum := ""
	if checksumAsset != nil {
		checksum, err = u.getChecksumForAsset(checksumAsset, assetNamePrefix)
		if err != nil {
			return nil, fmt.Errorf("failed to get checksum: %w", err)
		}
	}

	return &ReleaseInfo{
		Version:      version,
		Channel:      u.Channel,
		AssetURL:     assetURL,
		Checksum:     checksum,
		ReleaseNotes: release.GetBody(),
	}, nil
}

func (u *Updater) getChecksumForAsset(checksumAsset *github.ReleaseAsset, assetNamePrefix string) (string, error) {
	client := &http.Client{}
	resp, err := client.Get(checksumAsset.GetBrowserDownloadURL())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse checksum file
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, assetNamePrefix) {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				return fields[0], nil
			}
		}
	}

	return "", fmt.Errorf("checksum not found for %s", assetNamePrefix)
}

func getPlatformArch() string {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		return "x86_64"
	case "386":
		return "i386"
	default:
		return arch
	}
}

// // Mock for os.Exit
// var osExit = func(code int) {
// 	if os.Getenv("TEST_UPDATE_SUBPROCESS") == "1" {
// 		os.Exit(code)
// 	}
// 	panic("os.Exit called")
// }

// SaveRestartInfo saves the current command information for restart after update
func (u *Updater) SaveRestartInfo() error {
	info := &RestartInfo{
		Args:        os.Args,
		Environment: os.Environ(),
		WorkingDir:  getCurrentWorkingDir(),
		Timestamp:   time.Now(),
	}

	content, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return os.WriteFile(u.RestartPath, content, 0640)
}

// RestartAfterUpdate restarts the previously running command
func (u *Updater) RestartAfterUpdate() error {
	content, err := os.ReadFile(u.RestartPath)
	if err != nil {
		return fmt.Errorf("no restart information found: %w", err)
	}

	var info RestartInfo
	if err := json.Unmarshal(content, &info); err != nil {
		return fmt.Errorf("invalid restart information: %w", err)
	}

	// Clean up restart info file
	os.Remove(u.RestartPath)

	// Prepare the new process
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// Create new process with the same arguments and environment
	attr := &os.ProcAttr{
		Dir: info.WorkingDir,
		Env: info.Environment,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	// Start the new process
	process, err := os.StartProcess(executable, info.Args, attr)
	if err != nil {
		return err
	}

	// Detach from the new process
	if err := process.Release(); err != nil {
		return err
	}

	// Exit the current process
	// osExit(0)
	return nil
}

func getCurrentWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func (u *Updater) updateCache(release *ReleaseInfo) error {
	u.cache = &UpdateCache{
		LastCheck:     time.Now(),
		LatestVersion: release.Version.String(),
		Channel:       u.Channel,
		TTL:           24 * time.Hour,
	}

	content, err := json.Marshal(u.cache)
	if err != nil {
		return err
	}

	return os.WriteFile(u.CacheFile, content, 0640)
}

func (u *Updater) getTelemetryPath() string {
	return u.TelemetryPath
}
