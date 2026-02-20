package setup

import "time"

// StatusPayload is the JSON payload for setup status (GET /setup/status and 409 conflict).
type StatusPayload struct {
	Status     string     `json:"status"`
	Step       string     `json:"step,omitempty"`
	Error      string     `json:"error,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// NewStatusPayload builds a StatusPayload from individual fields.
// Zero times are omitted from the payload (nil pointers).
func NewStatusPayload(status, step, err string, startedAt, finishedAt time.Time) StatusPayload {
	p := StatusPayload{
		Status: status,
		Step:   step,
		Error:  err,
	}
	if !startedAt.IsZero() {
		t := startedAt
		p.StartedAt = &t
	}
	if !finishedAt.IsZero() {
		t := finishedAt
		p.FinishedAt = &t
	}
	return p
}
