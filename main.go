package main

import (
	"context"

	"api-gateway/cmd"

	"github.com/ihezebin/olympus/logger"
)

func main() {
	ctx := context.Background()
	if err := cmd.Run(ctx); err != nil {
		logger.Fatalf(ctx, "cmd run error: %v", err)
	}
}
