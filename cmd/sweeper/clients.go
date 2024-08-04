package main

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func newClients(ctx context.Context, config configuration) {
	logger.Init(ctx, config.logger)
}
