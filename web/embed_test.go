package web

import (
	"bytes"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestFrontendPages(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	index, err := assets.ReadFile("dist/index.html")
	if err != nil {
		t.Fatal(err)
	}

	for _, target := range []string{
		"/", "/index.html", "/login", "/register", "/dashboard",
		"/properties/1/rooms?month=2026-10", "/unknown-page",
	} {
		t.Run(target, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), index) {
				t.Fatalf("status = %d, response does not match embedded index.html", response.Code)
			}
			if !strings.HasPrefix(response.Header().Get("Content-Type"), "text/html") {
				t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
			}
			if response.Header().Get("Cache-Control") != "no-cache" {
				t.Fatalf("cache control = %q", response.Header().Get("Cache-Control"))
			}
		})
	}
}

func TestFrontendStaticResources(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	index, err := assets.ReadFile("dist/index.html")
	if err != nil {
		t.Fatal(err)
	}
	resources := regexp.MustCompile(`(?:src|href)="(/[^" ]+)"`).FindAllSubmatch(index, -1)
	if len(resources) == 0 {
		t.Fatal("built index.html contains no static resources")
	}
	for _, resource := range resources {
		target := string(resource[1])
		t.Run(target, func(t *testing.T) {
			want, err := fs.ReadFile(assets, "dist"+target)
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target+"?v=1", nil))
			if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), want) {
				t.Fatalf("status = %d, response does not match embedded asset", response.Code)
			}
			contentType := response.Header().Get("Content-Type")
			switch {
			case strings.HasSuffix(target, ".js"):
				if !strings.Contains(contentType, "javascript") {
					t.Fatalf("content type = %q", contentType)
				}
			case strings.HasSuffix(target, ".css"):
				if !strings.HasPrefix(contentType, "text/css") {
					t.Fatalf("content type = %q", contentType)
				}
			}
		})
	}
}

func TestFrontendMissingResourcesAndDirectories(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{
		"/assets/missing.js", "/assets/missing.css", "/missing.svg",
		"/assets/missing", "/assets", "/assets/", "/dist/index.html", "/package.json",
	} {
		t.Run(target, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", response.Code)
			}
		})
	}
}

func TestFrontendRequestMethods(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"/", "/login", "/favicon.svg"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodHead, target, nil))
		if response.Code != http.StatusOK || response.Body.Len() != 0 {
			t.Fatalf("HEAD %s: status = %d, body length = %d", target, response.Code, response.Body.Len())
		}
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(method, "/login", nil))
		if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != "GET, HEAD" {
			t.Fatalf("%s: status = %d, allow = %q", method, response.Code, response.Header().Get("Allow"))
		}
	}
}
