package updater

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPRequesterFetchSuccess(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response body"))
	}))
	defer server.Close()

	requester := HTTPRequester{}
	reader, err := requester.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if string(body) != "test response body" {
		t.Errorf("expected 'test response body', got '%s'", string(body))
	}
}

func TestHTTPRequesterFetchNon200(t *testing.T) {
	statusCodes := []int{
		http.StatusNotFound,
		http.StatusInternalServerError,
		http.StatusForbidden,
		http.StatusUnauthorized,
	}

	for _, statusCode := range statusCodes {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
			}))
			defer server.Close()

			requester := HTTPRequester{}
			_, err := requester.Fetch(server.URL)
			if err == nil {
				t.Fatalf("expected error for status %d", statusCode)
			}

			if !strings.Contains(err.Error(), "bad http status") {
				t.Errorf("expected 'bad http status' error, got: %v", err)
			}
		})
	}
}

func TestHTTPRequesterFetchInvalidURL(t *testing.T) {
	requester := HTTPRequester{}
	_, err := requester.Fetch("http://nonexistent.invalid/")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestHTTPRequesterFetchMalformedURL(t *testing.T) {
	requester := HTTPRequester{}
	_, err := requester.Fetch("not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for malformed URL")
	}
}

func TestRequesterInterface(t *testing.T) {
	// Verify HTTPRequester implements Requester interface
	var _ Requester = &HTTPRequester{}
	var _ Requester = &mockRequester{}
}

func TestHTTPRequesterFetchLargeResponse(t *testing.T) {
	// Test handling of larger responses
	largeContent := strings.Repeat("x", 1024*1024) // 1MB

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeContent))
	}))
	defer server.Close()

	requester := HTTPRequester{}
	reader, err := requester.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if len(body) != len(largeContent) {
		t.Errorf("expected %d bytes, got %d bytes", len(largeContent), len(body))
	}
}

func TestHTTPRequesterFetchEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write nothing
	}))
	defer server.Close()

	requester := HTTPRequester{}
	reader, err := requester.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if len(body) != 0 {
		t.Errorf("expected empty response, got %d bytes", len(body))
	}
}
