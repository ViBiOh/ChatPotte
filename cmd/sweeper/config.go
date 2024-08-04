package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/ChatPotte/discord"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

type configuration struct {
	logger  *logger.Config
	discord *discord.Config

	username *string
}

func newConfiguration() configuration {
	fs := flag.NewFlagSet("discord", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		logger:  logger.Flags(fs, "logger"),
		discord: discord.Flags(fs, ""),

		username: flags.New("username", "Username of messages to clean").DocPrefix("sweeper").String(fs, "", nil),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}
