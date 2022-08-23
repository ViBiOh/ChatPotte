package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/sha"
)

var cacheVersion = sha.New("vibioh/ChatPotte/1")[:8]

func cacheKey(prefix, content string) string {
	return fmt.Sprintf("%s:%s:%s", prefix, cacheVersion, content)
}

func (a App) SaveCustomID(ctx context.Context, redisApp redis.App, prefix, separator string, values ...string) (string, error) {
	content := strings.Join(values, separator)
	key := sha.New(content)
	return key, redisApp.Store(ctx, cacheKey(prefix, key), content, time.Hour)
}

func (a App) RestoreCustomID(ctx context.Context, redisApp redis.App, prefix, separator, customID string, statics []string) (string, error) {
	for _, static := range statics {
		if customID == static {
			return customID, nil
		}
	}

	content, err := redisApp.Load(ctx, cacheKey(prefix, customID))
	if err != nil {
		return customID, fmt.Errorf("load redis: %w", err)
	}

	return content, nil
}
