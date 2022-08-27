package discord

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/sha"
)

func cacheKey(prefix, content string) string {
	return fmt.Sprintf("%s:%s", prefix, content)
}

func SaveCustomID(ctx context.Context, redisApp redis.App, prefix string, values url.Values) (string, error) {
	content := values.Encode()
	key := sha.New(content)
	return key, redisApp.Store(ctx, cacheKey(prefix, key), content, time.Hour)
}

func RestoreCustomID(ctx context.Context, redisApp redis.App, prefix, customID string, statics []string) (url.Values, error) {
	for _, static := range statics {
		if customID == static {
			return url.ParseQuery(customID)
		}
	}

	content, err := redisApp.Load(ctx, cacheKey(prefix, customID))
	if err != nil {
		return nil, fmt.Errorf("load redis: %w", err)
	}

	return url.ParseQuery(content)
}
