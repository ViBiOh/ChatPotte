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
		defer close(messagesCh)

		for _, guild := range guilds {
			channels, err := discord.Channels(ctx, req, guild)
			logger.FatalfOnErr(ctx, err, "channels")

			for _, channel := range channels {
				if err := services.discord.Messages(ctx, req, channel.ID, messagesCh); err != nil {
					slog.Error("list messages", slog.String("guild", guild.Name), slog.String("channel", channel.Name), slog.Any("error", err))
				}
			}
		}
	}()

	var read, deleted uint

	for message := range messagesCh {
		read++

		if !message.Timestamp.Before(before) {
			continue
		}

		if shouldDelete(message, *config.userIDs, *config.usernames) {
			if err := services.discord.DeleteMessage(ctx, req, message); err != nil {
				slog.ErrorContext(ctx, "unable to delete delete message", slog.Any("error", err))
			} else {
				deleted++
			}
		}
	}

	slog.InfoContext(ctx, fmt.Sprintf("%d messages read, %d deleted", read, deleted))
}

func shouldDelete(message discord.Message, userIDs, usernames []string) bool {
	for _, userID := range userIDs {
		if message.Author.Bot && strings.Contains(message.Content, userID) {
			return true
		}
	}

	for _, username := range usernames {
		if strings.EqualFold(message.Author.Username, username) {
			return true
		}
	}

	return false
}
