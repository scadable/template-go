package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouter_MetricsEndpoint(t *testing.T) {
	router := NewRouter("test-service")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "# HELP") {
		t.Error("expected Prometheus metrics output")
	}
}

func TestRouter_DocsRedirect(t *testing.T) {
	router := NewRouter("test-service")

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301 redirect, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != "/docs/index.html" {
		t.Fatalf("expected redirect to /docs/index.html, got %s", location)
	}
}

func TestRouter_SwaggerHandler(t *testing.T) {
	router := NewRouter("test-service")

	req := httptest.NewRequest(http.MethodGet, "/docs/index.html", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	// We don't expect 404 because httpSwagger.WrapHandler should be attached
	if resp.StatusCode == http.StatusNotFound {
		t.Fatal("swagger route returned 404; is httpSwagger.WrapHandler configured correctly?")
	}
}

func TestRouter_RootRouteMounted(t *testing.T) {
	router := NewRouter("test-service")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		return // good
	}

	if resp.StatusCode == http.StatusNotFound {
		t.Fatal("expected root route to be mounted but got 404")
	}

	t.Fatalf("unexpected status code: %d", resp.StatusCode)
}
