package setup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agent-service-prototype/pkg/logger"
)

// SHA256Hex returns the SHA256 hash of data as a hex string.
func SHA256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// NormalizeChecksum strips an optional "sha256:" or "SHA256:" prefix from the expected value.
func NormalizeChecksum(expected string) string {
	const prefix = "sha256:"
	if len(expected) > len(prefix) && strings.EqualFold(expected[:len(prefix)], prefix) {
		return expected[len(prefix):]
	}
	return expected
}

// VerifyChecksums verifies that all files in checksums match their expected SHA256.
// Expected values may include an optional "sha256:" prefix.
func VerifyChecksums(baseDir string, checksums map[string]string) error {
	for relPath, expected := range checksums {
		data, err := os.ReadFile(filepath.Join(baseDir, relPath))
		if err != nil {
			return fmt.Errorf("failed to read %s for checksum: %w", relPath, err)
		}
		actual := SHA256Hex(data)
		expectedNorm := NormalizeChecksum(expected)
		if actual != expectedNorm {
			return fmt.Errorf("checksum mismatch for %s: expected=%s actual=%s", relPath, expectedNorm, actual)
		}
	}
	logger.Info().Int("files", len(checksums)).Msg("All checksums verified")
	return nil
}
