package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
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

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	logger.New(loggerConfig)

	ctx := context.Background()

	discordApp, err := discord.New(discordConfig, "", nil, nil)
	if err != nil {
		slog.Error("create discord", "err", err)
		os.Exit(1)
	}

	var commands map[string]discord.Command
	if err := json.Unmarshal([]byte(*configuration), &commands); err != nil {
		slog.Error("parse configuration", "err", err)
		os.Exit(1)
	}

	if err := discordApp.ConfigureCommands(ctx, commands); err != nil {
		slog.Error("configure command", "err", err)
		os.Exit(1)
	}
}
