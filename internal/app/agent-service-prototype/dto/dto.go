package dto

import "time"

type InstallationSuccess struct {
	Status          string  `json:"status"`
	SchemaVersion   string  `json:"schema_version"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type InstallationFailed struct {
	Status string `json:"status"`
	Step   string `json:"step"`
	Error  string `json:"error"`
}

type SetupStatus struct {
	Status     string     `json:"status"`
	Step       string     `json:"step,omitempty"`
	Error      string     `json:"error,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}
