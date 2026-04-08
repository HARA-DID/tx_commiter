package callback

import (
	"context"
)

type Result struct {
	EventID      string   `json:"event_id"`
	JobID        string   `json:"job_id"`
	EventType    string   `json:"event_type"`
	Success      bool     `json:"success"`
	TxHashes     []string `json:"tx_hashes,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
}

type Func func(ctx context.Context, result Result) error
