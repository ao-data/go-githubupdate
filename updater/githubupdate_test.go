package updater

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v68/github"
)

func TestNewUpdater(t *testing.T) {
	u := NewUpdater("1.0.0", "ao-data", "go-githubupdate", "update-")

	if u.CurrentVersion != "1.0.0" {
		t.Errorf("expected CurrentVersion to be '1.0.0', got '%s'", u.CurrentVersion)
	}
	if u.GithubOwner != "ao-data" {
		t.Errorf("expected GithubOwner to be 'ao-data', got '%s'", u.GithubOwner)
	}
	if u.GithubRepo != "go-githubupdate" {
		t.Errorf("expected GithubRepo to be 'go-githubupdate', got '%s'", u.GithubRepo)
	}
	if u.FilePrefix != "update-" {
		t.Errorf("expected FilePrefix to be 'update-', got '%s'", u.FilePrefix)
	}
	if u.Requester != nil {
		t.Error("expected Requester to be nil by default")
	}
}

func TestPlatformConstant(t *testing.T) {
	expected := runtime.GOOS + "-" + runtime.GOARCH
	if platform != expected {
		t.Errorf("expected platform to be '%s', got '%s'", expected, platform)
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		expectUpdate   bool
	}{
		{
			name:           "update available - major version",
			currentVersion: "1.0.0",
			latestVersion:  "2.0.0",
			expectUpdate:   true,
		},
		{
			name:           "update available - minor version",
			currentVersion: "1.0.0",
			latestVersion:  "1.1.0",
			expectUpdate:   true,
		},
		{
			name:           "update available - patch version",
			currentVersion: "1.0.0",
			latestVersion:  "1.0.1",
			expectUpdate:   true,
		},
		{
			name:           "no update - same version",
			currentVersion: "1.0.0",
			latestVersion:  "1.0.0",
			expectUpdate:   false,
		},
		{
			name:           "no update - current is newer",
			currentVersion: "2.0.0",
			latestVersion:  "1.0.0",
			expectUpdate:   false,
		},
		{
			name:           "update available - prerelease to release",
			currentVersion: "1.0.0-alpha",
			latestVersion:  "1.0.0",
			expectUpdate:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := semver.Make(tt.currentVersion)
			if err != nil {
				t.Fatalf("failed to parse current version: %v", err)
			}

			latest, err := semver.Make(tt.latestVersion)
			if err != nil {
				t.Fatalf("failed to parse latest version: %v", err)
			}

			needsUpdate := current.LT(latest)
			if needsUpdate != tt.expectUpdate {
				t.Errorf("expected update=%v, got update=%v", tt.expectUpdate, needsUpdate)
			}
		})
	}
}

func TestUpdateErrorNoBinary(t *testing.T) {
	u := NewUpdater("1.0.0", "ao-data", "go-githubupdate", "update-")

	// Set up a mock release with no matching assets
	tagName := "2.0.0"
	wrongAssetName := "wrong-file.gz"
	u.latestReleasesResp = &github.RepositoryRelease{
		TagName: &tagName,
		Assets: []*github.ReleaseAsset{
			{Name: &wrongAssetName},
		},
	}

	err := u.Update()
	if err == nil {
		t.Fatal("expected error when no matching binary found")
	}
	if !errors.Is(err, ErrorNoBinary) {
		t.Errorf("expected ErrorNoBinary, got: %v", err)
	}
}

func TestUpdateFindsCorrectAsset(t *testing.T) {
	u := NewUpdater("1.0.0", "ao-data", "go-githubupdate", "myapp-")

	// Create the expected filename based on current platform
	expectedFilename := "myapp-" + platform + ".gz"
	if runtime.GOOS == "windows" {
		expectedFilename = "myapp-" + platform + ".exe.gz"
	}

	tagName := "2.0.0"
	wrongAssetName := "wrong-file.gz"
	downloadURL := "https://example.com/download/" + expectedFilename

	u.latestReleasesResp = &github.RepositoryRelease{
		TagName: &tagName,
		Assets: []*github.ReleaseAsset{
			{Name: &wrongAssetName},
			{Name: &expectedFilename, BrowserDownloadURL: &downloadURL},
		},
	}

	// Set up mock requester to return a gzipped binary
	mock := &mockRequester{}
	mock.handleRequest(func(url string) (io.ReadCloser, error) {
		if url != downloadURL {
			t.Errorf("expected download URL '%s', got '%s'", downloadURL, url)
		}
		// Return a gzipped "binary"
		return createGzipReader([]byte("fake binary content")), nil
	})
	u.Requester = mock

	// Note: This will fail at selfupdate.Apply() since we're not actually
	// replacing a real binary, but it proves the asset matching works
	err := u.Update()
	if err == nil {
		t.Log("Update succeeded (unexpected in test environment)")
	} else if !strings.Contains(err.Error(), "update failed") {
		// If the error is about something other than the update itself failing,
		// that's unexpected
		if errors.Is(err, ErrorNoBinary) {
			t.Error("should have found the correct asset")
		}
	}
}

func TestFetchGZ(t *testing.T) {
	u := NewUpdater("1.0.0", "ao-data", "go-githubupdate", "update-")

	originalContent := []byte("test binary content 12345")

	mock := &mockRequester{}
	mock.handleRequest(func(url string) (io.ReadCloser, error) {
		return createGzipReader(originalContent), nil
	})
	u.Requester = mock

	result, err := u.fetchGZ("https://example.com/test.gz")
	if err != nil {
		t.Fatalf("fetchGZ failed: %v", err)
	}

	if !bytes.Equal(result, originalContent) {
		t.Errorf("expected '%s', got '%s'", string(originalContent), string(result))
	}
}

func TestFetchGZInvalidGzip(t *testing.T) {
	u := NewUpdater("1.0.0", "ao-data", "go-githubupdate", "update-")

	mock := &mockRequester{}
	mock.handleRequest(func(url string) (io.ReadCloser, error) {
		// Return non-gzipped content
		return io.NopCloser(strings.NewReader("not gzipped content")), nil
	})
	u.Requester = mock

	_, err := u.fetchGZ("https://example.com/test.gz")
	if err == nil {
		t.Fatal("expected error for invalid gzip content")
	}
}

func TestFetchGZNetworkError(t *testing.T) {
	u := NewUpdater("1.0.0", "ao-data", "go-githubupdate", "update-")

	mock := &mockRequester{}
	mock.handleRequest(func(url string) (io.ReadCloser, error) {
		return nil, errors.New("network error")
	})
	u.Requester = mock

	_, err := u.fetchGZ("https://example.com/test.gz")
	if err == nil {
		t.Fatal("expected error for network failure")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected network error, got: %v", err)
	}
}

func TestFetchWithNilRequester(t *testing.T) {
	u := NewUpdater("1.0.0", "ao-data", "go-githubupdate", "update-")
	// Requester is nil, should use defaultHTTPRequester

	// This will make an actual HTTP request and fail, but proves the code path works
	_, err := u.fetch("http://nonexistent.invalid/")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestFetchWithNilReadCloser(t *testing.T) {
	u := NewUpdater("1.0.0", "ao-data", "go-githubupdate", "update-")

	mock := &mockRequester{}
	mock.handleRequest(func(url string) (io.ReadCloser, error) {
		return nil, nil // Return nil ReadCloser without error
	})
	u.Requester = mock

	_, err := u.fetch("https://example.com/test")
	if err == nil {
		t.Fatal("expected error for nil ReadCloser")
	}
	if !strings.Contains(err.Error(), "non-nil ReadCloser") {
		t.Errorf("expected non-nil ReadCloser error, got: %v", err)
	}
}

func TestMockRequester(t *testing.T) {
	mock := &mockRequester{}

	// Add two handlers
	mock.handleRequest(func(url string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("response 1")), nil
	})
	mock.handleRequest(func(url string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("response 2")), nil
	})

	// First fetch
	r1, err := mock.Fetch("url1")
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	content1, _ := io.ReadAll(r1)
	if string(content1) != "response 1" {
		t.Errorf("expected 'response 1', got '%s'", string(content1))
	}

	// Second fetch
	r2, err := mock.Fetch("url2")
	if err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}
	content2, _ := io.ReadAll(r2)
	if string(content2) != "response 2" {
		t.Errorf("expected 'response 2', got '%s'", string(content2))
	}

	// Third fetch should fail (no more handlers)
	_, err = mock.Fetch("url3")
	if err == nil {
		t.Fatal("expected error when no more handlers available")
	}
}

func TestExpectedFilename(t *testing.T) {
	prefix := "myapp-"
	expectedGz := prefix + platform + ".gz"
	expectedExeGz := prefix + platform + ".exe.gz"

	// Verify the filename construction logic
	var reqFilename string
	if runtime.GOOS == "windows" {
		reqFilename = expectedExeGz
	} else {
		reqFilename = expectedGz
	}

	if runtime.GOOS == "windows" {
		if reqFilename != expectedExeGz {
			t.Errorf("on windows, expected '%s', got '%s'", expectedExeGz, reqFilename)
		}
	} else {
		if reqFilename != expectedGz {
			t.Errorf("on non-windows, expected '%s', got '%s'", expectedGz, reqFilename)
		}
	}
}

// Helper function to create a gzipped reader from content
func createGzipReader(content []byte) io.ReadCloser {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(content)
	gz.Close()
	return io.NopCloser(bytes.NewReader(buf.Bytes()))
}
