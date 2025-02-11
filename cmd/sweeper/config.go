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

	userIDs   *[]string
	usernames *[]string
}

func newConfiguration() configuration {
	fs := flag.NewFlagSet("discord", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		logger:  logger.Flags(fs, "logger"),
		discord: discord.Flags(fs, ""),

		userIDs:   flags.New("user", "User ID to clean").DocPrefix("sweeper").StringSlice(fs, nil, nil),
		usernames: flags.New("username", "Username to clean").DocPrefix("sweeper").StringSlice(fs, nil, nil),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}
