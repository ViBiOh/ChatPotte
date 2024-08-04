package main

import (
	"fmt"

	"github.com/ViBiOh/ChatPotte/discord"
)

type services struct {
	discord discord.Service
}

func newServices(config configuration) (services, error) {
	var output services
	var err error

	output.discord, err = discord.New(config.discord, "", nil, nil)
	if err != nil {
		return output, fmt.Errorf("discord: %w", err)
	}

	return output, nil
}
