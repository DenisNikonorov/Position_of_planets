package redis

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"planets/internal/app/config"

	"github.com/redis/go-redis/v9"
)

// servicePrefix — префикс ключей этого сервиса в общем Redis.
// Как и в методичке, префиксы с разделителем не дают ключам
// разных сервисов пересекаться.
const servicePrefix = "planets."

// Client — обёртка над Redis, используется для блеклиста токенов.
type Client struct {
	client *redis.Client
}

// New подключается к Redis и проверяет соединение пингом:
// если хранилище недоступно, приложение не стартует.
func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Username: cfg.User,
		Password: cfg.Password,
		DB:       0,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
