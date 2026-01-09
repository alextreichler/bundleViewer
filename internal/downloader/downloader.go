package downloader

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetBinaryPath returns the expected path for a downloaded binary
func GetBinaryPath(dataDir, version, arch string) string {
	// Version often comes as "v25.3.1 - hash", we need just "25.3.1" or "v25.3.1"
	// The URL uses "25.3.1" (no v)
	cleanVer := cleanVersion(version)
	return filepath.Join(dataDir, "binaries", fmt.Sprintf("redpanda-%s-%s", cleanVer, arch))
}

func cleanVersion(v string) string {
	// "v25.3.1 - ..." -> "25.3.1"
	parts := strings.Split(v, " ")
	ver := parts[0]
	return strings.TrimPrefix(ver, "v")
}

// DownloadRedpanda downloads and extracts the Redpanda binary.
// Returns the path to the executable.
func DownloadRedpanda(dataDir, version, arch string) (string, error) {
	cleanVer := cleanVersion(version)
	
	// Normalize arch for URL (amd64 is standard)
	if arch == "x86_64" {
		arch = "amd64"
	}

	binDir := filepath.Join(dataDir, "binaries")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create binaries dir: %w", err)
	}

	targetPath := filepath.Join(binDir, fmt.Sprintf("redpanda-%s-%s", cleanVer, arch))
	
	// Check if already exists
	if _, err := os.Stat(targetPath); err == nil {
		slog.Info("Binary already exists", "path", targetPath)
		return targetPath, nil
	}

	// URL construction
	// https://vectorized-public.s3.us-west-2.amazonaws.com/releases/redpanda/25.3.1/redpanda-25.3.1-amd64.tar.gz
	url := fmt.Sprintf("https://vectorized-public.s3.us-west-2.amazonaws.com/releases/redpanda/%s/redpanda-%s-%s.tar.gz", cleanVer, cleanVer, arch)
	
	slog.Info("Downloading Redpanda binary", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Stream extraction
	if err := extractTarGz(resp.Body, targetPath); err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		return "", fmt.Errorf("failed to chmod: %w", err)
	}

	slog.Info("Download and extraction complete", "path", targetPath)
	return targetPath, nil
}

func extractTarGz(r io.Reader, targetPath string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for the main binary.
		// Usually in tarball it's: redpanda-25.3.1-amd64/bin/redpanda OR just redpanda?
		// Or maybe etc/..., bin/...
		// We only want the 'redpanda' binary.
		
		// The tarball structure usually has a top level dir.
		// Let's look for any file ending in "/bin/redpanda" or just "redpanda" if it's top level.
		if strings.HasSuffix(header.Name, "/bin/redpanda") || header.Name == "redpanda" || strings.HasSuffix(header.Name, "/redpanda") {
			// Found it. Extract to targetPath (which is a FILE, not a dir in our design here)
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			return nil
		}
	}

	return fmt.Errorf("redpanda binary not found in tarball")
}
