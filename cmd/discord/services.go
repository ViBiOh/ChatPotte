package main

import (
	"fmt"

	"github.com/ViBiOh/ChatPotte/discord"
)

type services struct {
	discord discord.Service
}

func newServices(config configuration) (services, error) {
	discordService, err := discord.New(config.discord, "", nil, nil)
	if err != nil {
		return services{}, fmt.Errorf("discord: %w", err)
	}

	return services{
		discord: discordService,
	}, nil
}
