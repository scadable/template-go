package routes

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"template-go/pkg/logger"
	"testing"
)

func TestRootRoutes_HelloWorld(t *testing.T) {
	router := RootRoutes()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("expected content-type 'text/plain', got '%s'", contentType)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Hello, World!" {
		t.Errorf("expected body 'Hello, World!', got '%s'", string(body))
	}
}

func TestRootRoutes_NotFound(t *testing.T) {
	router := RootRoutes()

	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown route, got %d", resp.StatusCode)
	}
}

// mockFailingWriter simulates an http.ResponseWriter where Write always fails.
type mockFailingWriter struct {
	header http.Header
}

// Header returns a non-nil header map.
func (m *mockFailingWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

// WriteHeader is a no-op that satisfies the interface.
func (m *mockFailingWriter) WriteHeader(statusCode int) {}

// Write always returns an error to test the handler's error path.
func (m *mockFailingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("simulated write error")
}

func TestHelloWorld_WriteError(t *testing.T) {
	// Arrange: Initialize the logger to prevent a nil pointer panic.
	logger.Init()

	// Arrange: Create a request; its details don't matter for this test.
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Arrange: Use the mock writer that is designed to fail.
	rec := &mockFailingWriter{}

	// Act: Call the handler directly. Its internal call to w.Write() will now fail,
	// executing the `if err != nil` block.
	helloWorld(rec, req)

	// Assert: We can still check that the header was set correctly,
	// even though the write operation failed.
	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("expected content-type 'text/plain', got '%s'", contentType)
	}
}
