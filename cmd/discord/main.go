package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ViBiOh/ChatPotte/discord"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("DiscordConfigure", flag.ExitOnError)

	loggerConfig := logger.Flags(fs, "logger")
	discordConfig := discord.Flags(fs, "")
	inputFile := flags.String(fs, "", "discord", "Input", "JSON file containing commands definition", "", nil)

	logger.Fatal(fs.Parse(os.Args[1:]))

	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	discordApp, err := discord.New(discordConfig, "", nil)
	logger.Fatal(err)

	file, err := os.Open(*inputFile)
	if err != nil {
		logger.Fatal(fmt.Errorf("unable to open input file: %s", err))
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.Error("unable to close input file: %s", err)
		}
	}()

	var commands map[string]discord.Command
	if err := json.NewDecoder(file).Decode(&commands); err != nil {
		logger.Fatal(fmt.Errorf("unable to parse input file: %s", err))
	}

	logger.Fatal(discordApp.ConfigureCommands(commands))
}
