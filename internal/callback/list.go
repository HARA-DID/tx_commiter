package callback

import (
	"context"
	"log"
)

func LogResult(_ context.Context, result Result) error {
	if result.Success {
		log.Printf("[CALLBACK] Event %s (Job %s) succeeded. TxHashes: %v", result.EventID, result.JobID, result.TxHashes)
	} else {
		log.Printf("[CALLBACK] Event %s (Job %s) FAILED: %s", result.EventID, result.JobID, result.ErrorMessage)
	}
	return nil
}

func NoOp(_ context.Context, _ Result) error {
	return nil
}
