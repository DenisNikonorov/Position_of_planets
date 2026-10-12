package handler

import (
	"errors"
	"net/http"
	"strings"

	"planets/internal/app/auth"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

// jwtPrefix — префикс заголовка Authorization, как в RFC 6750.
const jwtPrefix = "Bearer "

// bearerToken достаёт токен из заголовка Authorization.
// Второе значение — был ли заголовок вообще.
func bearerToken(ctx *gin.Context) (string, bool) {
	header := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(header, jwtPrefix) {
		return "", false
	}
	return strings.TrimPrefix(header, jwtPrefix), true
}

// WithAuthCheck — middleware методов создателя. Пропускает запрос только
// если присланный токен валиден и не отозван:
//
//	нет заголовка или невалидный токен → 401;
//	токен в блеклисте Redis (после logout) → 403;
//	сбой хранилища → 500.
//
// После проверки claims кладутся в контекст, дальше их читает auth.
func (h *Handler) WithAuthCheck(ctx *gin.Context) {
	raw, ok := bearerToken(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	switch err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), raw); {
	case err == nil:
		// токен в блеклисте — он отозван, подпись всё ещё корректна
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	case !errors.Is(err, goredis.Nil):
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	claims, err := auth.ParseToken(h.JWT, raw)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	auth.SetUser(ctx, claims)
}

// OptionalAuthCheck — middleware публичных GET. Токен не обязателен:
// если он валиден, пользователь проставляется в контексте (тогда в ответе
// посчитаются liked и is_owner), если нет — запрос обрабатывается как гость.
func (h *Handler) OptionalAuthCheck(ctx *gin.Context) {
	raw, ok := bearerToken(ctx)
	if ok {
		if err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), raw); errors.Is(err, goredis.Nil) {
			if claims, err := auth.ParseToken(h.JWT, raw); err == nil {
				auth.SetUser(ctx, claims)
			}
		}
	}
	ctx.Next()
}
