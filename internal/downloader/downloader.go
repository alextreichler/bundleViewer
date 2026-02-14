package downloader

import (
	"archive/tar"
	"archive/zip"
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

// FetchAndExtractBundle downloads a bundle from a URL and extracts it to a temporary directory.
// It supports .zip, .tar.gz, and .tgz formats.
func FetchAndExtractBundle(url, dataDir string) (string, error) {
	slog.Info("Fetching bundle from URL", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download bundle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download bundle: status %d", resp.StatusCode)
	}

	// Create a temporary directory for extraction
	extractDir := filepath.Join(dataDir, "extracted_bundle")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// Clean extraction directory if it already exists
	if entries, err := os.ReadDir(extractDir); err == nil && len(entries) > 0 {
		slog.Info("Cleaning existing extraction directory")
		os.RemoveAll(extractDir)
		os.MkdirAll(extractDir, 0755)
	}

	// Detect format and extract
	if strings.HasSuffix(url, ".zip") {
		return extractDir, extractZip(resp.Body, extractDir)
	} else if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
		return extractDir, extractBundleTarGz(resp.Body, extractDir)
	}

	return "", fmt.Errorf("unsupported bundle format (expected .zip, .tar.gz, or .tgz)")
}

func extractBundleTarGz(r io.Reader, dest string) error {
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

		target := filepath.Join(dest, header.Name)
		
		// Prevent ZipSlip / Path Traversal
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) && target != dest {
			continue // Skip suspicious paths
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func extractZip(r io.Reader, dest string) error {
	// zip.NewReader needs an io.ReaderAt and the size.
	// This means we have to buffer to disk or memory first.
	// Let's buffer to a temporary file.
	tmpFile, err := os.CreateTemp("", "bundle-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	size, err := io.Copy(tmpFile, r)
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(tmpFile, size)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		target := filepath.Join(dest, f.Name)

		// Prevent ZipSlip / Path Traversal
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) && target != dest {
			continue // Skip suspicious paths
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}

		outFile.Close()
		rc.Close()
	}

	return nil
}
