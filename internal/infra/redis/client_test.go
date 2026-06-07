package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"

	"speakmate/internal/config"
)

func TestOpenClientPingsRedisWhenEnabled(t *testing.T) {
	server := miniredis.RunT(t)

	client, err := OpenClient(context.Background(), config.RedisConfig{
		Enabled:               true,
		Addr:                  server.Addr(),
		DB:                    0,
		ConnectTimeoutSeconds: 1,
	})
	if err != nil {
		t.Fatalf("OpenClient returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Ping returned error after OpenClient: %v", err)
	}
}

func TestOpenClientReturnsClearErrorWhenRedisUnavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := OpenClient(ctx, config.RedisConfig{
		Enabled:               true,
		Addr:                  "127.0.0.1:1",
		DB:                    0,
		ConnectTimeoutSeconds: 1,
	})

	if !errors.Is(err, ErrRedisUnavailable) {
		t.Fatalf("OpenClient error = %v, want ErrRedisUnavailable", err)
	}
}
