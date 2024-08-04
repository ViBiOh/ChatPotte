package discord

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

type Guild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func CurrentUser(ctx context.Context, req request.Request) (User, error) {
	resp, err := req.Path("/users/@me").Method(http.MethodGet).Send(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("get: %w", err)
	}

	return httpjson.Read[User](resp)
}

func Guilds(ctx context.Context, req request.Request) ([]Guild, error) {
	resp, err := req.Path("/users/@me/guilds").Method(http.MethodGet).Send(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	return httpjson.Read[[]Guild](resp)
}

func Channels(ctx context.Context, req request.Request, guild Guild) ([]Channel, error) {
	resp, err := req.Path("/guilds/%s/channels", guild.ID).Method(http.MethodGet).Send(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	return httpjson.Read[[]Channel](resp)
}
