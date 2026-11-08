package ds

import "github.com/golang-jwt/jwt/v5"

// JWTClaims — содержимое access-токена. RegisteredClaims даёт стандартные
// по RFC 7519 поля (exp, iat, iss), остальные — данные пользователя из БД.
// Роли (is_moderator) кладём в токен, чтобы не читать БД на каждом запросе.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      uint   `json:"user_id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

// LoginUserRequest — тело POST /api/users/auth.
type LoginUserRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse — тело ответа POST /api/users/auth.
// Клиент передаёт AccessToken в заголовке Authorization: Bearer <token>.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}
