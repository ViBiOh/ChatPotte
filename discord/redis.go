package discord

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/hash"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
)

func cacheKey(prefix, content string) string {
	return fmt.Sprintf("%s:%s", prefix, content)
}

func SaveCustomID(ctx context.Context, redisApp redis.Client, prefix string, values url.Values) (string, error) {
	content := values.Encode()
	key := hash.String(content)
	return key, redisApp.Store(ctx, cacheKey(prefix, key), content, time.Hour)
}

func RestoreCustomID(ctx context.Context, redisApp redis.Client, prefix, customID string, statics []string) (url.Values, error) {
	if slices.Contains(statics, customID) {
		return url.ParseQuery(customID)
	}

	content, err := redisApp.Load(ctx, cacheKey(prefix, customID))
	if err != nil {
		return nil, fmt.Errorf("load redis: %w", err)
	}

	return url.ParseQuery(string(content))
}
