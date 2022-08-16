package discord

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// ConfigureCommands with the API
func (a App) ConfigureCommands(commands map[string]Command) error {
	if len(a.applicationID) == 0 {
		return nil
	}

	ctx := context.Background()

	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "applications.commands.update")

	resp, err := discordRequest.Method(http.MethodPost).Path("/oauth2/token").BasicAuth(a.clientID, a.clientSecret).Form(ctx, data)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	content := make(map[string]any)
	if err := httpjson.Read(resp, &content); err != nil {
		return fmt.Errorf("read oauth token: %w", err)
	}

	bearer := content["access_token"].(string)
	rootURL := fmt.Sprintf("/applications/%s", a.applicationID)

	for name, command := range commands {
		for _, registerURL := range getRegisterURLs(command) {
			absoluteURL := rootURL + registerURL
			logger.WithField("command", name).Info("Configuring with URL `%s`", absoluteURL)

			if _, err := discordRequest.Method(http.MethodPost).Path(absoluteURL).Header("Authorization", fmt.Sprintf("Bearer %s", bearer)).StreamJSON(ctx, command); err != nil {
				return fmt.Errorf("configure `%s` command: %w", name, err)
			}
		}

		logger.Info("Command `%s` configured!", name)
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
