package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func makeTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(content)), Mode: 0755}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func TestFetchLatestReleaseFromURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(release{TagName: "v1.2.3"})
	}))
	defer srv.Close()

	rel, err := fetchLatestReleaseFromURL(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if rel.TagName != "v1.2.3" {
		t.Errorf("want v1.2.3, got %s", rel.TagName)
	}
}

func TestFetchLatestReleaseFromURL_NetworkError(t *testing.T) {
	_, err := fetchLatestReleaseFromURL("http://127.0.0.1:0/invalid")
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestFetchLatestReleaseFromURL_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := fetchLatestReleaseFromURL(srv.URL)
	if err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestFetchLatestReleaseFromURL_EmptyTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(release{TagName: ""})
	}))
	defer srv.Close()

	_, err := fetchLatestReleaseFromURL(srv.URL)
	if err == nil {
		t.Error("expected error for empty tag_name")
	}
}

func TestDownloadAndExtract(t *testing.T) {
	archive := makeTarGz(t, "ccsm", []byte("binary-content"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.Write(archive)
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	binPath, err := downloadAndExtract(srv.URL, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if binPath != filepath.Join(tmpDir, "ccsm") {
		t.Errorf("unexpected path: %s", binPath)
	}
	content, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "binary-content" {
		t.Errorf("unexpected content: %s", content)
	}
	info, _ := os.Stat(binPath)
	if info.Mode()&0111 == 0 {
		t.Error("binary is not executable")
	}
}

func TestDownloadAndExtract_BinaryNotFound(t *testing.T) {
	archive := makeTarGz(t, "other-file", []byte("data"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(archive)
	}))
	defer srv.Close()

	_, err := downloadAndExtract(srv.URL, t.TempDir())
	if err == nil {
		t.Error("expected error when ccsm binary absent from archive")
	}
}

func TestDownloadAndExtract_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := downloadAndExtract(srv.URL, t.TempDir())
	if err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestReplaceBinary_RequiresRoot(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("test must run as non-root")
	}
	src := filepath.Join(t.TempDir(), "ccsm")
	if err := os.WriteFile(src, []byte("bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := replaceBinary(src, "/usr/local/bin/ccsm"); err == nil {
		t.Error("expected error when not root")
	}
}
