package upgrade

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	githubAPI      = "https://api.github.com/repos/nemethk/claude-code-session-manager/releases/latest"
	releaseBaseURL = "https://github.com/nemethk/claude-code-session-manager/releases/download"
	installPath    = "/usr/local/bin/ccsm"
)

type release struct {
	TagName string `json:"tag_name"`
}

func Run(currentVersion string) error {
	fmt.Println("Fetching latest release...")
	rel, err := fetchLatestReleaseFromURL(githubAPI)
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}

	fmt.Printf("Latest version: %s\n", rel.TagName)

	// Normalise versions for comparison (strip leading 'v')
	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")
	if current != "dev" && current == latest {
		fmt.Println("Already up to date.")
		return nil
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos != "linux" && goos != "darwin" {
		return fmt.Errorf("unsupported OS: %s", goos)
	}
	if goarch != "amd64" && goarch != "arm64" {
		return fmt.Errorf("unsupported architecture: %s", goarch)
	}

	fmt.Printf("Detected: %s %s\n", goos, goarch)

	tmpDir, err := os.MkdirTemp("", "ccsm-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archiveName := fmt.Sprintf("ccsm_%s_%s.tar.gz", goos, goarch)
	downloadURL := fmt.Sprintf("%s/%s/%s", releaseBaseURL, rel.TagName, archiveName)

	fmt.Printf("Downloading from: %s\n", downloadURL)
	binaryPath, err := downloadAndExtract(downloadURL, tmpDir)
	if err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	fmt.Printf("Installing to %s...\n", installPath)
	if err := replaceBinary(binaryPath, installPath); err != nil {
		return err
	}

	out, err := exec.Command(installPath, "--version").Output()
	if err != nil {
		return fmt.Errorf("failed to verify: %w", err)
	}
	fmt.Printf("Upgrade complete: %s", string(out))
	return nil
}

func fetchLatestReleaseFromURL(url string) (*release, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("no release tag found")
	}
	return &rel, nil
}

func downloadAndExtract(url, destDir string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		if hdr.Name != "ccsm" {
			continue
		}

		dest := filepath.Join(destDir, "ccsm")
		f, err := os.Create(dest)
		if err != nil {
			return "", fmt.Errorf("failed to create file: %w", err)
		}
		if _, err = io.Copy(f, tr); err != nil {
			f.Close()
			return "", fmt.Errorf("failed to extract binary: %w", err)
		}
		f.Close()

		if err := os.Chmod(dest, 0755); err != nil {
			return "", fmt.Errorf("failed to chmod: %w", err)
		}
		return dest, nil
	}

	return "", fmt.Errorf("ccsm binary not found in archive")
}

func replaceBinary(src, dest string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("upgrade requires root: run 'sudo ccsm upgrade'")
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination binary: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	if err := out.Chmod(0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}
	return nil
}
