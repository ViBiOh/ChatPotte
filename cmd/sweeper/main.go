package main

import (
	"context"
	"log/slog"
	"runtime"
	"strings"
	"sync"
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

	before := time.Now().AddDate(0, -3, 0)

	for _, guild := range guilds {
		channels, err := discord.Channels(ctx, req, guild)
		logger.FatalfOnErr(ctx, err, "channels")

		var wg sync.WaitGroup
		messagesCh := make(chan discord.Message, runtime.NumCPU())

		go func() {
			wg.Wait()
			close(messagesCh)
		}()

		for _, channel := range channels {
			wg.Add(1)
			go func() {
				defer wg.Done()

				if err := services.discord.Messages(ctx, req, channel.ID, messagesCh); err != nil {
					slog.Error("list messages", slog.Any("error", err))
				}
			}()
		}

		for message := range messagesCh {
			if message.Timestamp.Before(before) && strings.EqualFold(message.Author.Username, *config.username) {
				slog.Info("Deleting message", slog.String("message", message.String()))
				logger.FatalfOnErr(ctx, services.discord.DeleteMessage(ctx, req, message), "delete")
			}
		}
	}
}
