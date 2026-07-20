package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func prettyPrint(model any) (string, error) {
	b, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to MarshalIndent the model: %w", err)
	}
	return string(b), nil
}

func getContext() context.Context {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return ctx
}
