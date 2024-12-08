package updater

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v35/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		repo          string
		channel       UpdateChannel
		expectedError bool
	}{
		{
			name:          "Valid initialization",
			version:       "1.0.0",
			repo:          "martient/bifrost-backups",
			channel:       StableChannel,
			expectedError: false,
		},
		{
			name:          "Invalid repo format",
			version:       "1.0.0",
			repo:          "invalid-repo",
			channel:       StableChannel,
			expectedError: false, // Error will occur during fetch, not initialization
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater, err := New(tt.version, tt.repo, tt.channel)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, updater)
			assert.Equal(t, tt.version, updater.CurrentVersion)
			assert.Equal(t, tt.repo, updater.GithubRepo)
			assert.Equal(t, tt.channel, updater.Channel)
		})
	}
}

func TestFetchLatestRelease(t *testing.T) {
	// Create a mock binary server
	mockBinaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mock binary content"))
	}))
	defer mockBinaryServer.Close()

	// Create a mock checksum server
	mockChecksumServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
		assetName := fmt.Sprintf("bifrost-backups_%s.tar.gz", platform)
		mockBinary := []byte("mock binary content")
		mockChecksum := fmt.Sprintf("%x", sha256.Sum256(mockBinary))
		w.Write([]byte(fmt.Sprintf("%s %s", mockChecksum, assetName)))
	}))
	defer mockChecksumServer.Close()

	// Create a mock GitHub server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
		assetName := fmt.Sprintf("bifrost-backups_%s.tar.gz", platform)

		releases := []*github.RepositoryRelease{
			{
				TagName:    github.String("2.0.0"),
				Draft:      github.Bool(false),
				Prerelease: github.Bool(false),
				Assets: []*github.ReleaseAsset{
					createMockAsset(assetName, mockBinaryServer.URL),
					createMockAsset("bifrost-backups_checksums.txt", mockChecksumServer.URL),
				},
			},
			{
				TagName:    github.String("2.1.0-beta.1"),
				Draft:      github.Bool(false),
				Prerelease: github.Bool(true),
				Assets: []*github.ReleaseAsset{
					createMockAsset(assetName, mockBinaryServer.URL),
					createMockAsset("bifrost-backups_checksums.txt", mockChecksumServer.URL),
				},
			},
		}
		json.NewEncoder(w).Encode(releases)
	}))
	defer mockServer.Close()

	// Create a new client that uses the mock server
	client := github.NewClient(&http.Client{
		Transport: &mockTransport{mockServer},
	})

	tests := []struct {
		name            string
		channel         UpdateChannel
		expectedVersion string
	}{
		{
			name:            "Stable channel gets latest stable",
			channel:         StableChannel,
			expectedVersion: "2.0.0",
		},
		{
			name:            "Beta channel gets latest beta",
			channel:         BetaChannel,
			expectedVersion: "2.1.0-beta.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater, err := New("1.0.0", "martient/bifrost-backups", tt.channel)
			require.NoError(t, err)
			updater.client = client

			release, err := updater.fetchLatestRelease()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVersion, release.Version.String())
		})
	}
}

func TestUpdate(t *testing.T) {
	if os.Getenv("TEST_UPDATE_SUBPROCESS") == "1" {
		return // Skip the test in the subprocess
	}

	// Mock os.Exit
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()
	// var exitCode int
	// osExit = func(code int) {
	// 	exitCode = code
	// 	panic("os.Exit called")
	// }

	// t.Run("Successful update", func(t *testing.T) {
	// 	defer func() {
	// 		if r := recover(); r != nil {
	// 			if r != "os.Exit called" {
	// 				t.Fatalf("unexpected panic: %v", r)
	// 			}
	// 			assert.Equal(t, 0, exitCode)
	// 		}
	// 	}()

	// 	// Get the current executable path
	// 	currentExe, err := os.Executable()
	// 	require.NoError(t, err)

	// 	// Read the current executable as our mock binary
	// 	mockBinary, err := os.ReadFile(currentExe)
	// 	require.NoError(t, err)
	// 	mockChecksum := fmt.Sprintf("%x", sha256.Sum256(mockBinary))

	// 	// Create a mock server for binary download
	// 	mockBinaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		w.Write(mockBinary)
	// 	}))
	// 	defer mockBinaryServer.Close()

	// 	// Create a temporary directory for the test
	// 	tmpDir, err := os.MkdirTemp("", "updater_test")
	// 	require.NoError(t, err)
	// 	defer os.RemoveAll(tmpDir)

	// 	// Create a mock executable by copying the current executable
	// 	exePath := filepath.Join(tmpDir, "bifrost-backup")
	// 	err = os.WriteFile(exePath, mockBinary, 0755)
	// 	require.NoError(t, err)

	// 	// Generate a key pair for testing
	// 	pubKey, privKey, err := ed25519.GenerateKey(nil)
	// 	require.NoError(t, err)

	// 	// Set up the updater
	// 	u, err := New("1.0.0", "martient/bifrost-backups", StableChannel)
	// 	require.NoError(t, err)
	// 	u.PublicKey = pubKey

	// 	// Sign the binary
	// 	signature := ed25519.Sign(privKey, mockBinary)

	// 	// Create release info
	// 	release := &ReleaseInfo{
	// 		Version:      semver.MustParse("2.0.0"),
	// 		Channel:      StableChannel,
	// 		AssetURL:     mockBinaryServer.URL,
	// 		Checksum:     mockChecksum,
	// 		ReleaseNotes: "Test release notes",
	// 		Signature:    fmt.Sprintf("%x", signature),
	// 	}

	// 	// Set up environment for subprocess
	// 	os.Setenv("TEST_UPDATE_SUBPROCESS", "1")
	// 	defer os.Unsetenv("TEST_UPDATE_SUBPROCESS")

	// 	// Perform the update
	// 	err = u.Update(release)
	// 	require.NoError(t, err)

	// 	// Verify the binary was updated
	// 	updatedBinary, err := os.ReadFile(exePath)
	// 	require.NoError(t, err)
	// 	assert.Equal(t, mockBinary, updatedBinary)
	// })

	t.Run("Failed checksum verification", func(t *testing.T) {
		// Create a mock server for binary download
		mockBinaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("mock binary content"))
		}))
		defer mockBinaryServer.Close()

		// Create a temporary directory for the test
		tmpDir, err := os.MkdirTemp("", "updater_test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create a mock executable
		exePath := filepath.Join(tmpDir, "bifrost-backup")
		err = os.WriteFile(exePath, []byte("old binary"), 0755)
		require.NoError(t, err)

		// Set up the updater
		u, err := New("1.0.0", "martient/bifrost-backups", StableChannel)
		require.NoError(t, err)

		// Create release info with invalid checksum
		release := &ReleaseInfo{
			Version:      semver.MustParse("2.0.0"),
			Channel:      StableChannel,
			AssetURL:     mockBinaryServer.URL,
			Checksum:     "invalid-checksum",
			ReleaseNotes: "Test release notes",
		}

		// Perform the update
		err = u.Update(release)
		assert.Error(t, err)
	})
}

func TestRestartAfterUpdate(t *testing.T) {
	t.Run("Valid restart info", func(t *testing.T) {
		// Create a temporary directory for the test
		tmpDir, err := os.MkdirTemp("", "updater_test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create updater instance
		u, err := New("1.0.0", "martient/bifrost-backups", StableChannel)
		require.NoError(t, err)
		u.RestartPath = filepath.Join(tmpDir, "restart.json")

		// Create a mock restart info file
		restartInfo := struct {
			Args        []string  `json:"args"`
			Environment []string  `json:"environment"`
			WorkingDir  string    `json:"working_dir"`
			Timestamp   time.Time `json:"timestamp"`
		}{
			Args:        []string{"test-binary", "--arg1", "--arg2"},
			Environment: []string{"TEST_ENV=value"},
			WorkingDir:  "/test/dir",
			Timestamp:   time.Now(),
		}

		restartInfoBytes, err := json.Marshal(restartInfo)
		require.NoError(t, err)

		err = os.WriteFile(u.RestartPath, restartInfoBytes, 0644)
		require.NoError(t, err)

		// Mock os.Exit to prevent test from actually exiting
		oldOsExit := osExit
		defer func() { osExit = oldOsExit }()

		exitCode := 0
		osExit = func(code int) {
			exitCode = code
			panic("os.Exit called")
		}

		// Test RestartAfterUpdate
		defer func() {
			if r := recover(); r != nil {
				assert.Equal(t, "os.Exit called", r)
				assert.Equal(t, 0, exitCode)
			}
		}()

		u.RestartAfterUpdate()
	})

	t.Run("Missing restart info", func(t *testing.T) {
		// Create a temporary directory for the test
		tmpDir, err := os.MkdirTemp("", "updater_test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create updater instance
		u, err := New("1.0.0", "martient/bifrost-backups", StableChannel)
		require.NoError(t, err)
		u.RestartPath = filepath.Join(tmpDir, "restart.json")

		// Test RestartAfterUpdate
		err = u.RestartAfterUpdate()
		assert.Error(t, err)
	})
}

func TestGetChecksumForAsset(t *testing.T) {
	mockBinary := []byte("mock binary content")
	mockChecksum := fmt.Sprintf("%x", sha256.Sum256(mockBinary))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checksums := fmt.Sprintf("%s bifrost-backups_darwin_arm64.tar.gz\n%s bifrost-backups_linux_amd64.tar.gz",
			mockChecksum,
			"f5e9c6d7b8a3b2c1d4e5f6g7h8i9j0k1")
		w.Write([]byte(checksums))
	}))
	defer mockServer.Close()

	tests := []struct {
		name           string
		assetName      string
		expectedHash   string
		expectError    bool
		checksumServer *httptest.Server
	}{
		{
			name:           "Valid checksum",
			assetName:      "bifrost-backups_darwin_arm64.tar.gz",
			expectedHash:   mockChecksum,
			expectError:    false,
			checksumServer: mockServer,
		},
		{
			name:           "Missing asset in checksums",
			assetName:      "bifrost-backups_windows_amd64.tar.gz",
			expectedHash:   "",
			expectError:    true,
			checksumServer: mockServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := &github.ReleaseAsset{
				Name:               github.String("bifrost-backups_checksums.txt"),
				BrowserDownloadURL: github.String(tt.checksumServer.URL),
			}

			u, err := New("1.0.0", "martient/bifrost-backups", StableChannel)
			require.NoError(t, err)

			hash, err := u.getChecksumForAsset(asset, tt.assetName)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedHash, hash)
		})
	}
}

// Helper functions

func createMockAsset(name, url string) *github.ReleaseAsset {
	return &github.ReleaseAsset{
		Name:               github.String(name),
		BrowserDownloadURL: github.String(url),
	}
}

type mockTransport struct {
	mockServer *httptest.Server
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace GitHub API URL with mock server URL
	req.URL.Scheme = "http"
	req.URL.Host = t.mockServer.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(req)
}

// Mock for os.Exit
var osExit = os.Exit
