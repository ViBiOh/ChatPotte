package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"

	"github.com/ViBiOh/ChatPotte/discord"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("discord", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	loggerConfig := logger.Flags(fs, "logger")
	discordConfig := discord.Flags(fs, "")
	configuration := flags.New("", "Configuration of commands, as JSON string").Prefix("commands").String(fs, "", nil)

	_ = fs.Parse(os.Args[1:])

	ctx := context.Background()

	logger.Init(ctx, loggerConfig)

	discordApp, err := discord.New(discordConfig, "", nil, nil)
	logger.FatalfOnErr(ctx, err, "create discord")

	var commands map[string]discord.Command
	if err := json.Unmarshal([]byte(*configuration), &commands); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "parse configuration", slog.Any("error", err))
		os.Exit(1)
	}

	if err := discordApp.ConfigureCommands(ctx, commands); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "configure command", slog.Any("error", err))
		os.Exit(1)
	}
}
