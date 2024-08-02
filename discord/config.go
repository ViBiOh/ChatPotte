package discord

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func (s Service) ConfigureCommands(ctx context.Context, commands map[string]Command) error {
	if len(s.applicationID) == 0 {
		return nil
	}

	req, err := s.SigninClient(ctx)
	if err != nil {
		return fmt.Errorf("signin: %w", err)
	}

	rootURL := fmt.Sprintf("/applications/%s", s.applicationID)

	for name, command := range commands {
		for _, registerURL := range getRegisterURLs(command) {
			absoluteURL := rootURL + registerURL

		configure:
			if resp, err := req.Method(http.MethodPost).Path(absoluteURL).StreamJSON(ctx, command); err != nil {
				if resp.StatusCode == http.StatusTooManyRequests {
					if resetAt, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
						duration := time.Until(time.Unix(resetAt, 0))

						slog.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf("Rate-limited, waiting %d before retrying...", duration), slog.String("url", absoluteURL))
						time.Sleep(duration)

						goto configure
					}
				}

				return fmt.Errorf("configure `%s` command for url `%s`: %w", name, registerURL, err)
			}
		}

		slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("Command `%s` configured!", name))
	}

	return nil
}

func (s Service) SigninClient(ctx context.Context) (request.Request, error) {
	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "applications.commands.update")

	resp, err := discordRequest.Method(http.MethodPost).Path("/oauth2/token").BasicAuth(s.clientID, s.clientSecret).Form(ctx, data)
	if err != nil {
		return discordRequest, fmt.Errorf("get token: %w", err)
	}

	content := make(map[string]any)
	if err := httpjson.Read(resp, &content); err != nil {
		return discordRequest, fmt.Errorf("read oauth token: %w", err)
	}

	bearer := content["access_token"].(string)

	return discordRequest.Header("authorization", fmt.Sprintf("Bearer %s", bearer)), nil
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
