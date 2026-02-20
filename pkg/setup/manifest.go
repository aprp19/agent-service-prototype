package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"agent-service-prototype/pkg/logger"
)

// BaselinePath unmarshals from either a JSON string ("path/to/baseline.sql")
// or an object with a "file" or "path" key.
type BaselinePath string

func (b *BaselinePath) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*b = BaselinePath(s)
		return nil
	}
	var obj struct {
		File string `json:"file"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	if obj.File != "" {
		*b = BaselinePath(obj.File)
	} else {
		*b = BaselinePath(obj.Path)
	}
	return nil
}

// Manifest describes the db-bundle structure.
type Manifest struct {
	Baseline   BaselinePath `json:"baseline"`
	Migrations []Migration  `json:"migrations"`
	Checks     struct {
		Smoke string `json:"smoke"`
	} `json:"checks"`
}

// Migration describes a single migration file.
type Migration struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	File        string `json:"file"`
	Transaction bool   `json:"transaction"`
}

// LoadManifest reads and parses manifest.json from baseDir.
func LoadManifest(baseDir string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest.json: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %w", err)
	}
	logger.Info().
		Str("baseline", string(m.Baseline)).
		Int("migrations", len(m.Migrations)).
		Str("smoke", m.Checks.Smoke).
		Msg("Manifest parsed")
	return &m, nil
}
