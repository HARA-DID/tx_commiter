package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	streamName := os.Getenv("STREAM_NAME")
	if streamName == "" {
		streamName = "did-events"
	}

	groupName := os.Getenv("GROUP_NAME")
	if groupName == "" {
		groupName = "worker-group"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse redis url: %v\n", err)
		os.Exit(1)
	}

	client := redis.NewClient(opts)
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("Analyzing PEL (Pending Entries List) for Group: %q on Stream: %q\n", groupName, streamName)
	fmt.Printf("Displaying unacknowledged messages...\n\n")

	pendings, err := client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamName,
		Group:  groupName,
		Start:  "-",
		End:    "+",
		Count:  50, // Display top 50 pending messages
	}).Result()

	if err != nil {
		if err == redis.Nil || len(pendings) == 0 {
			fmt.Println("No pending messages found in PEL. Everything is ACKed!")
			return
		}
		fmt.Fprintf(os.Stderr, "XPENDING error: %v\n", err)
		os.Exit(1)
	}

	for _, p := range pendings {
		msgs, err := client.XRange(ctx, streamName, p.ID, p.ID).Result()
		var payload string
		if err == nil && len(msgs) > 0 {
			raw, _ := json.MarshalIndent(msgs[0].Values, "    ", "  ")
			payload = string(raw)
		} else {
			payload = "  [Payload no longer exists in stream (likely trimmed)]"
		}

		fmt.Printf("── Pending Message %s ───────────────────\n", p.ID)
		fmt.Printf("  Consumer   : %s\n", p.Consumer)
		fmt.Printf("  Idle Time  : %s\n", p.Idle)
		fmt.Printf("  Deliveries : %d (Redelivered if > 1)\n", p.RetryCount)
		fmt.Printf("  Payload    :\n%s\n\n", payload)
	}

	if len(pendings) >= 50 {
		fmt.Println("... (capped at 50 results)")
	}
}
