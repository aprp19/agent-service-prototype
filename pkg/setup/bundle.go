package setup

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"agent-service-prototype/pkg/logger"
)

// DownloadBundle downloads a file from url to dest.
func DownloadBundle(ctx context.Context, url, dest string) error {
	logger.Info().Str("url", url).Str("dest", dest).Msg("Downloading bundle")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download bundle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from bundle URL", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write bundle: %w", err)
	}

	logger.Info().Str("dest", dest).Msg("Bundle downloaded")
	return nil
}

// ExtractZip extracts src zip to dest directory with zip-slip protection.
func ExtractZip(src, dest string) error {
	logger.Info().Str("src", src).Str("dest", dest).Msg("Extracting bundle")

	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	destClean := filepath.Clean(dest)
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		targetClean := filepath.Clean(target)
		if targetClean != destClean && !strings.HasPrefix(targetClean, destClean+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("failed to create dir for %s: %w", f.Name, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
		}

		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create %s: %w", target, err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return fmt.Errorf("failed to extract %s: %w", f.Name, err)
		}

		out.Close()
		rc.Close()
	}

	logger.Info().Str("dest", dest).Msg("Bundle extracted")
	return nil
}

// ResolveBaseDir locates the directory containing manifest.json.
// Handles zips with files at root or inside a single top-level folder.
func ResolveBaseDir(extractDir string) (string, error) {
	if _, err := os.Stat(filepath.Join(extractDir, "manifest.json")); err == nil {
		return extractDir, nil
	}

	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return "", fmt.Errorf("failed to read extract dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			candidate := filepath.Join(extractDir, e.Name())
			if _, err := os.Stat(filepath.Join(candidate, "manifest.json")); err == nil {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("manifest.json not found in extracted bundle")
}

// LoadChecksums reads and parses checksums.json from baseDir.
func LoadChecksums(baseDir string) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, "checksums.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read checksums.json: %w", err)
	}
	var checksums map[string]string
	if err := json.Unmarshal(data, &checksums); err != nil {
		return nil, fmt.Errorf("failed to parse checksums.json: %w", err)
	}
	return checksums, nil
}
