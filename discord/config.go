package discord

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
)

func (s Service) ConfigureCommands(ctx context.Context, commands map[string]Command) error {
	if len(s.applicationID) == 0 {
		return nil
	}

	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "applications.commands.update")

	resp, err := discordRequest.Method(http.MethodPost).Path("/oauth2/token").BasicAuth(s.clientID, s.clientSecret).Form(ctx, data)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	content := make(map[string]any)
	if err := httpjson.Read(resp, &content); err != nil {
		return fmt.Errorf("read oauth token: %w", err)
	}

	bearer := content["access_token"].(string)
	rootURL := fmt.Sprintf("/applications/%s", s.applicationID)

	for name, command := range commands {
		for _, registerURL := range getRegisterURLs(command) {
			absoluteURL := rootURL + registerURL

		configure:
			if resp, err := discordRequest.Method(http.MethodPost).Path(absoluteURL).Header("Authorization", fmt.Sprintf("Bearer %s", bearer)).StreamJSON(ctx, command); err != nil {
				if resp.StatusCode == http.StatusTooManyRequests {
					slog.LogAttrs(ctx, slog.LevelWarn, "Rate-limited, waiting 5 seconds before retrying...", slog.String("url", absoluteURL))
					time.Sleep(time.Second * 5)

					goto configure
				}

				return fmt.Errorf("configure `%s` command for url `%s`: %w", name, registerURL, err)
			}
		}

		slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("Command `%s` configured!", name))
	}

	return nil
}

func getRegisterURLs(command Command) []string {
	if len(command.Guilds) == 0 {
		return []string{"/commands"}
	}

	urls := make([]string, len(command.Guilds))

	for i, guild := range command.Guilds {
		urls[i] = fmt.Sprintf("/guilds/%s/commands", guild)
	}

	return urls
}
