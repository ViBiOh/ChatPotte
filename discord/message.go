package discord

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

type Message struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Timestamp time.Time `json:"timestamp"`
	Author    User      `json:"author"`
	Content   string    `json:"content"`
	Embeds    []Embed   `json:"Embeds"`
}

func (m Message) String() string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("[%s] %s: %s", m.Timestamp.Format(time.RFC3339), m.Author.Username, m.Content))

	for _, embed := range m.Embeds {
		output.WriteString(fmt.Sprintf(", %s - %s", embed.Title, embed.Description))
	}

	return output.String()
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}

func (s Service) Messages(ctx context.Context, req request.Request, channelID string, output chan<- Message) error {
	baseURL := fmt.Sprintf("/channels/%s/messages?limit=100", channelID)
	nextURL := baseURL

	for {
	retry:
		resp, err := req.Path(nextURL).Send(ctx, nil)
		if err != nil {
			if IsRetryable(ctx, resp) {
				goto retry
			}

			return fmt.Errorf("list: %w", err)
		}

		messages, err := httpjson.Read[[]Message](resp)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		for _, message := range messages {
			output <- message
		}

		if len(messages) == 0 {
			return nil
		}

		nextURL = baseURL + "&before=" + messages[len(messages)-1].ID
	}
}

func (s Service) DeleteMessage(ctx context.Context, req request.Request, message Message) error {
retry:
	resp, err := req.Path("/channels/%s/messages/%s", message.ChannelID, message.ID).Method(http.MethodDelete).Send(ctx, nil)
	if err != nil {
		if IsRetryable(ctx, resp) {
			goto retry
		}

		return fmt.Errorf("delete: %w", err)
	}

	if err := request.DiscardBody(resp.Body); err != nil {
		return fmt.Errorf("discard: %w", err)
	}

	return nil
}
