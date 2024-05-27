package cache

import (
	"context"
	"testing"
)

func TestRedis(t *testing.T) {
	ctx := context.Background()
	if err := InitRedisCache(ctx, []string{"127.0.0.1:6379"}, "root"); err != nil {
		t.Fatal(err)
	}

	client := RedisCacheClient()

	if err := client.Set(ctx, "key", "value", 0).Err(); err != nil {
		t.Fatal(err)
	}
}
