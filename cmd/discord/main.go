package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/ViBiOh/ChatPotte/discord"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func main() {
	config := newConfiguration()

	ctx := context.Background()

	newClients(ctx, config)

	services, err := newServices(config)
	logger.FatalfOnErr(ctx, err, "services")

	var commands map[string]discord.Command
	if err := json.Unmarshal([]byte(*config.configuration), &commands); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "parse configuration", slog.Any("error", err))
		os.Exit(1)
	}

	if err := services.discord.ConfigureCommands(ctx, commands); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "configure command", slog.Any("error", err))
		os.Exit(1)
	}
}
