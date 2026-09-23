package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSPAFileServer_FallsBackToIndexForRoutes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>shell</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "robots.txt"), []byte("User-agent: *"), 0o644); err != nil {
		t.Fatal(err)
	}
	handler := spaFileServer(dir)

	cases := []struct {
		path     string
		wantCode int
		wantBody string
	}{
		{"/dashboard", http.StatusOK, "shell"},
		{"/docs?chapter=invariants", http.StatusOK, "shell"},
		{"/robots.txt", http.StatusOK, "User-agent"},
		{"/assets/missing.js", http.StatusNotFound, ""},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rec.Code != tc.wantCode {
			t.Errorf("%s: status = %d, want %d", tc.path, rec.Code, tc.wantCode)
		}
		if !strings.Contains(rec.Body.String(), tc.wantBody) {
			t.Errorf("%s: body = %q, want it to contain %q", tc.path, rec.Body.String(), tc.wantBody)
		}
	}
}
