package eventlog

import "time"

// EventLog represents a structured event or operation record sent from fab to DC.
type EventLog struct {
	ID        int64     `json:"id"`
	FabCode   string    `json:"fab_code"`
	Type      string    `json:"type"` // e.g. deploy, alert, info, error
	Message   string    `json:"message"`
	Severity  string    `json:"severity"` // info, warning, critical
	CreatedAt time.Time `json:"created_at"`
}
