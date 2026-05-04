package callback

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
)

type Result struct {
	EventID      string   `json:"event_id"`
	JobID        string   `json:"job_id"`
	EventType    string   `json:"event_type"`
	Success      bool     `json:"success"`
	TxHashes     []string    `json:"tx_hashes,omitempty"`
	Logs         []*types.Log `json:"logs,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

type Func func(ctx context.Context, result Result) error
