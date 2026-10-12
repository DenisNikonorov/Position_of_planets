package auth

import (
	"errors"
	"time"

	"planets/internal/app/config"
	"planets/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Ключи, под которыми данные из токена лежат в контексте запроса.
// Их проставляет middleware, обработчики только читают.
const (
	userIDKey      = "auth.user_id"
	loginKey       = "auth.login"
	isModeratorKey = "auth.is_moderator"
)

// ErrInvalidToken — токен отсутствует, плохо сформирован или просрочен.
var ErrInvalidToken = errors.New("invalid token")

// CreateToken подписывает access-токен для пользователя.
// Срок жизни и секрет берутся из конфигурации (.env).
func CreateToken(cfg config.JWTConfig, user *ds.User) (string, error) {
	now := time.Now()
	claims := ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
		UserID:      user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// ParseToken проверяет подпись и срок жизни токена и возвращает claims.
func ParseToken(cfg config.JWTConfig, raw string) (*ds.JWTClaims, error) {
	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// SetUser кладёт данные из токена в контекст запроса.
// Вызывается middleware после успешной проверки токена.
func SetUser(ctx *gin.Context, claims *ds.JWTClaims) {
	ctx.Set(userIDKey, claims.UserID)
	ctx.Set(loginKey, claims.Login)
	ctx.Set(isModeratorKey, claims.IsModerator)
}

// CurrentUserID — id пользователя из токена. Для гостя (токена не было)
// возвращает 0, поэтому подделать создателя записи через запрос нельзя.
func CurrentUserID(ctx *gin.Context) uint {
	return ctx.GetUint(userIDKey)
}

// CurrentLogin — логин пользователя из токена, пустая строка у гостя.
func CurrentLogin(ctx *gin.Context) string {
	value, _ := ctx.Get(loginKey)
	login, _ := value.(string)
	return login
}

// CurrentIsModerator — признак модератора из токена, false у гостя.
func CurrentIsModerator(ctx *gin.Context) bool {
	value, _ := ctx.Get(isModeratorKey)
	moderator, _ := value.(bool)
	return moderator
}
