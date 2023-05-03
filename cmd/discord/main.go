package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ViBiOh/ChatPotte/discord"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("discord", flag.ExitOnError)

	loggerConfig := logger.Flags(fs, "logger")
	discordConfig := discord.Flags(fs, "")
	configuration := flags.New("commands", "Configuration of commands, as JSON string").Prefix("commands").String(fs, "", nil)

	logger.Fatal(fs.Parse(os.Args[1:]))

	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	ctx := context.Background()

	discordApp, err := discord.New(discordConfig, "", nil, nil)
	logger.Fatal(err)

	var commands map[string]discord.Command
	if err := json.Unmarshal([]byte(*configuration), &commands); err != nil {
		logger.Fatal(fmt.Errorf("parse configuration: %w", err))
	}

	logger.Fatal(discordApp.ConfigureCommands(ctx, commands))
}
