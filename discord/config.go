package discord

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func (s Service) ConfigureCommands(ctx context.Context, commands map[string]Command) error {
	if len(s.applicationID) == 0 {
		return nil
	}

	req, err := s.SigninClient(ctx, "applications.commands.update")
	if err != nil {
		return fmt.Errorf("signin: %w", err)
	}

	rootURL := fmt.Sprintf("/applications/%s", s.applicationID)

	for name, command := range commands {
		for _, registerURL := range getRegisterURLs(command) {
			absoluteURL := rootURL + registerURL

		configure:
			if resp, err := req.Method(http.MethodPost).Path(absoluteURL).StreamJSON(ctx, command); err != nil {
				if IsRetryable(ctx, resp) {
					goto configure
				}

				return fmt.Errorf("configure `%s` command for url `%s`: %w", name, registerURL, err)
			}
		}

		slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("Command `%s` configured!", name))
	}

	return nil
}

func (s Service) SigninClient(ctx context.Context, scopes ...string) (request.Request, error) {
	if len(s.botToken) != 0 {
		return discordRequest.Header("Authorization", fmt.Sprintf("Bot %s", s.botToken)), nil
	}

	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("scope", strings.Join(scopes, " "))

	resp, err := discordRequest.Method(http.MethodPost).Path("/oauth2/token").BasicAuth(s.clientID, s.clientSecret).Form(ctx, data)
	if err != nil {
		return discordRequest, fmt.Errorf("get token: %w", err)
	}

	content, err := httpjson.Read[map[string]any](resp)
	if err != nil {
		return discordRequest, fmt.Errorf("read oauth token: %w", err)
	}

	tokenType, _ := content["token_type"].(string)
	token, _ := content["access_token"].(string)

	return discordRequest.Header("Authorization", fmt.Sprintf("%s %s", tokenType, token)), nil
}

func IsRetryable(ctx context.Context, resp *http.Response) bool {
	if resp.StatusCode != http.StatusTooManyRequests {
		return false
	}

	if duration, err := strconv.ParseInt(resp.Header.Get("Retry-after"), 10, 64); err == nil {
		slog.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf("Rate-limited, waiting %ds before retrying...", duration), slog.String("method", resp.Request.Method), slog.String("url", resp.Request.URL.Path))
		time.Sleep(time.Duration(duration) * time.Second)

		return true
	}

	return false
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
