package redis

import (
	"context"
	"time"
)

const jwtPrefix = "jwt."

// getJWTKey собирает ключ блеклиста: токен целиком, поэтому
// совпадений между пользователями не бывает.
func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

// WriteJWTToBlacklist кладёт выданный токен в блеклист на время его жизни.
// Смысл хранить токен есть только до exp: после истечения он и так невалиден,
// а запись можно не тратить.
func (c *Client) WriteJWTToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	return c.client.Set(ctx, getJWTKey(token), "revoked", ttl).Err()
}

// CheckJWTInBlacklist проверяет, не был ли токен отозван через logout.
// nil — токен в блеклисте, redis.Nil — токена нет (действует),
// любая другая ошибка — сбой хранилища.
func (c *Client) CheckJWTInBlacklist(ctx context.Context, token string) error {
	return c.client.Get(ctx, getJWTKey(token)).Err()
}
