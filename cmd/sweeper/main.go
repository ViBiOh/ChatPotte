package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/ViBiOh/ChatPotte/discord"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

type Guild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	config := newConfiguration()

	ctx := context.Background()

	newClients(ctx, config)

	services, err := newServices(config)
	logger.FatalfOnErr(ctx, err, "services")

	req, err := services.discord.SigninClient(ctx, "guilds", "identify", "messages.read", "bot")
	logger.FatalfOnErr(ctx, err, "signin")

	guilds, err := discord.Guilds(ctx, req)
	logger.FatalfOnErr(ctx, err, "guilds")

	before := time.Now().AddDate(0, -2, 0)

	messagesCh := make(chan discord.Message, runtime.NumCPU())

	go func() {
		for _, guild := range guilds {
			channels, err := discord.Channels(ctx, req, guild)
			logger.FatalfOnErr(ctx, err, "channels")

			for _, channel := range channels {
				if err := services.discord.Messages(ctx, req, channel.ID, messagesCh); err != nil {
					slog.Error("list messages", slog.Any("error", err))
				}
			}
		}

		close(messagesCh)
	}()

	for message := range messagesCh {
		if !message.Timestamp.Before(before) {
			continue
		}

		if shouldDelete(*config.currentUser, message, *config.usernames) {
			fmt.Println(message.String())
			logger.FatalfOnErr(ctx, services.discord.DeleteMessage(ctx, req, message), "delete")
		}
	}
}

func shouldDelete(currentUser string, message discord.Message, usernames []string) bool {
	if message.Author.Bot && strings.Contains(message.Content, currentUser) {
		return true
	}

	for _, username := range usernames {
		if strings.EqualFold(message.Author.Username, username) {
			return true
		}
	}

	return false
}
