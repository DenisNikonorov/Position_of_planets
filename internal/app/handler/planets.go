package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"planets/internal/app/auth"
	"planets/internal/app/config"
	"planets/internal/app/ds"
	"planets/internal/app/redis"
	"planets/internal/app/repository"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWT        config.JWTConfig
}

func NewHandler(r *repository.Repository, rdb *redis.Client, jwtCfg config.JWTConfig) *Handler {
	return &Handler{
		Repository: r,
		Redis:      rdb,
		JWT:        jwtCfg,
	}
}

// RegisterHandler регистрирует маршруты веб-сервиса. Все адреса
// начинаются с /api, домен услуг — /planets, документация — /swagger.
//
// Публичны аутентификация и чтение (GET): список, лента, регистрация и вход.
// Методы создателя (создание, черновик, публикация, лайк, удаление, logout)
// закрыты WithAuthCheck — без валидного токена они недоступны.
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		// Чтение без токена: токен, если приложен, учитывается
		// в liked/is_owner через OptionalAuthCheck.
		public := api.Group("", h.OptionalAuthCheck)
		public.GET("/planets", h.GetAllPlanets)
		public.GET("/planets/feed", h.GetFeed)

		// Аутентификация.
		api.POST("/users", h.RegisterUser)
		api.POST("/users/auth", h.AuthUser)

		// Методы создателя: только с Authorization: Bearer <access_token>.
		creator := api.Group("", h.WithAuthCheck)
		creator.POST("/planets", h.AddPlanet)
		creator.GET("/planets/draft", h.GetDraft)
		creator.PUT("/planets/:id", h.PublishPlanet)
		creator.POST("/planets/:id/like", h.LikePlanet)
		creator.DELETE("/planets/:id", h.DeletePlanet)
		creator.POST("/users/deauth", h.DeauthUser)
	}
}

// fail отдаёт пустое тело и стандартный код состояния: сообщения об
// ошибках клиенту не нужны, достаточно кода.
func (h *Handler) fail(ctx *gin.Context, code int) {
	ctx.Status(code)
}

// planetID разбирает :id из пути. Некорректный идентификатор — 400.
func planetID(ctx *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

// GetAllPlanets — GET /api/planets?search=&maxDist=
// Список опубликованных планет с фильтрацией на бэкенде.
// Доступно без токена: чтение публичное.
//
// @Summary      Список планет
// @Description  Опубликованные планеты с фильтрами. Работает и без токена: у гостя liked и is_owner равны 0.
// @Tags         planets
// @Produce      json
// @Param        search  query string false "Поиск по названию (ILIKE)"
// @Param        maxDist query number false "Максимальное расстояние от Солнца"
// @Success      200 {array} ds.PlanetSerializer
// @Failure      500 {string} string ""
// @Router       /planets [get]
func (h *Handler) GetAllPlanets(ctx *gin.Context) {
	var maxDist *float64
	if raw := ctx.Query("maxDist"); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			maxDist = &parsed
		}
	}

	planets, err := h.Repository.GetPlanets(ctx.Query("search"), maxDist, auth.CurrentUserID(ctx))
	if err != nil {
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, planets)
}

// GetFeed — GET /api/planets/feed?id=&next=true
// Лента опубликованных планет, листание по кругу. Доступно без токена.
//
// @Summary      Лента планет
// @Description  Одна планета из ленты. Без id — первая; с id и next=true — следующая после id (по кругу).
// @Tags         planets
// @Produce      json
// @Param        id   query integer false "Идентификатор текущей планеты"
// @Param        next query boolean false "Взять следующую после id"
// @Success      200 {object} ds.PlanetSerializer
// @Failure      400 {string} string ""
// @Failure      404 {string} string ""
// @Router       /planets/feed [get]
func (h *Handler) GetFeed(ctx *gin.Context) {
	var id *uint
	if raw := ctx.Query("id"); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			h.fail(ctx, http.StatusBadRequest)
			return
		}
		value := uint(parsed)
		id = &value
	}

	planet, err := h.Repository.GetFeed(id, ctx.Query("next") == "true", auth.CurrentUserID(ctx))
	if err != nil {
		h.fail(ctx, http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, planet)
}

// GetDraft — GET /api/planets/draft
// Черновик текущего пользователя, идентификатор не передаётся.
//
// @Summary      Поля черновика
// @Description  Последний черновик пользователя из токена. Требуется Authorization: Bearer.
// @Tags         planets
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ds.PlanetSerializer
// @Failure      401 {string} string ""
// @Failure      403 {string} string ""
// @Failure      404 {string} string ""
// @Router       /planets/draft [get]
func (h *Handler) GetDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetUserDraft(auth.CurrentUserID(ctx))
	if err != nil {
		h.fail(ctx, http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

// AddPlanet — POST /api/planets (этап 1 публикации).
// multipart/form-data: name и файлы под ключами image_url, url_video.
// Создаёт черновик с пустыми бизнес-полями и складывает медиафайлы в MinIO.
// Пользователь заполняется из токена, а не из тела запроса.
// Возвращает 201 и JSON с заполненными именем и ссылками на медиа.
//
// @Summary      Создать планету (черновик)
// @Description  Первый этап публикации. Владелец берётся из токена, поэтому подделать user_id нельзя. Отсутствующие файлы заменяются дефолтными.
// @Tags         planets
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        name      formData string true  "Название планеты"
// @Param        image_url formData file   false "Изображение (иначе дефолтное)"
// @Param        url_video formData file   false "Видео (иначе дефолтное)"
// @Success      201 {object} ds.PlanetSerializer
// @Failure      400 {string} string ""
// @Failure      401 {string} string ""
// @Failure      403 {string} string ""
// @Failure      500 {string} string ""
// @Router       /planets [post]
func (h *Handler) AddPlanet(ctx *gin.Context) {
	if err := ctx.Request.ParseMultipartForm(2 << 20); err != nil {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	var req ds.CreatePlanetRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	imageName, err := h.Repository.UploadMedia(fileHeader(ctx, "image_url"))
	if err != nil {
		h.fail(ctx, http.StatusBadRequest)
		return
	}
	if imageName == "" {
		imageName = repository.DefaultPhotoName
	}

	videoName, err := h.Repository.UploadMedia(fileHeader(ctx, "url_video"))
	if err != nil {
		h.fail(ctx, http.StatusBadRequest)
		return
	}
	if videoName == "" {
		videoName = repository.DefaultVideoName
	}

	planet, err := h.Repository.CreatePlanet(auth.CurrentUserID(ctx), req.Name, imageName, videoName)
	if err != nil {
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, planet)
}

// fileHeader достаёт файл из формы. Отсутствие файла — не ошибка:
// в этом случае подставится дефолтное медиа.
func fileHeader(ctx *gin.Context, field string) *multipart.FileHeader {
	header, err := ctx.FormFile(field)
	if err != nil {
		return nil
	}
	return header
}

// PublishPlanet — PUT /api/planets/:id (этап 2 публикации).
// Тело: JSON с бизнес-полями. Статус меняет только бизнес-логика:
// переход разрешён из draft, обратно в черновик вернуться нельзя.
// Гость получает 401, чужая или уже опубликованная запись — 404.
//
// @Summary      Опубликовать планету
// @Description  Второй этап публикации: заполняет бизнес-поля черновика и переводит его в published. Изменить чужую запись нельзя.
// @Tags         planets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path integer true "Идентификатор планеты"
// @Param        body body ds.PublishPlanetRequest true "Бизнес-поля планеты"
// @Success      200 {object} ds.PlanetSerializer
// @Failure      400 {string} string ""
// @Failure      401 {string} string ""
// @Failure      403 {string} string ""
// @Failure      404 {string} string ""
// @Failure      500 {string} string ""
// @Router       /planets/{id} [put]
func (h *Handler) PublishPlanet(ctx *gin.Context) {
	id, ok := planetID(ctx)
	if !ok {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	var req ds.PublishPlanetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	planet, err := h.Repository.PublishPlanet(auth.CurrentUserID(ctx), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.fail(ctx, http.StatusNotFound)
			return
		}
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, planet)
}

// LikePlanet — POST /api/planets/:id/like
// Тело: {"like": 1} ставит лайк, {"like": 0} отменяет.
//
// @Summary      Поставить или снять лайк
// @Description  Лайк ставит авторизованный пользователь на опубликованную планету.
// @Tags         planets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path integer true "Идентификатор планеты"
// @Param        body body ds.LikeRequest true "1 — поставить, 0 — снять"
// @Success      200 {object} ds.PlanetSerializer
// @Failure      400 {string} string ""
// @Failure      401 {string} string ""
// @Failure      403 {string} string ""
// @Failure      404 {string} string ""
// @Failure      500 {string} string ""
// @Router       /planets/{id}/like [post]
func (h *Handler) LikePlanet(ctx *gin.Context) {
	id, ok := planetID(ctx)
	if !ok {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	var req ds.LikeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	planet, err := h.Repository.SetLike(auth.CurrentUserID(ctx), id, *req.Like)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.fail(ctx, http.StatusNotFound)
			return
		}
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, planet)
}

// DeletePlanet — DELETE /api/planets/:id
// Только soft delete и только для планеты текущего пользователя.
//
// @Summary      Удалить планету
// @Description  Мягкое удаление: запись остаётся в БД, но исчезает из выдачи. Удалить можно только свою.
// @Tags         planets
// @Produce      json
// @Security     BearerAuth
// @Param        id path integer true "Идентификатор планеты"
// @Success      200 {object} ds.PlanetSerializer
// @Failure      400 {string} string ""
// @Failure      401 {string} string ""
// @Failure      403 {string} string ""
// @Failure      404 {string} string ""
// @Router       /planets/{id} [delete]
func (h *Handler) DeletePlanet(ctx *gin.Context) {
	id, ok := planetID(ctx)
	if !ok {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	planet, err := h.Repository.SoftDeletePlanet(auth.CurrentUserID(ctx), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.fail(ctx, http.StatusForbidden)
			return
		}
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, planet)
}

// RegisterUser — POST /api/users. Доступно без токена.
//
// @Summary      Регистрация пользователя
// @Description  Создаёт пользователя с bcrypt-хешем пароля. Логин должен быть свободен.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body ds.RegisterUserRequest true "Данные нового пользователя"
// @Success      201 {object} ds.UserSerializer
// @Failure      400 {string} string ""
// @Failure      403 {string} string ""
// @Failure      500 {string} string ""
// @Router       /users [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req ds.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	user, err := h.Repository.CreateUser(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrForbidden) {
			h.fail(ctx, http.StatusForbidden)
			return
		}
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// AuthUser — POST /api/users/auth
// Аутентификация: сверяет пароль с хешем из БД и выдаёт JWT.
// Токен передаётся дальше в заголовке Authorization: Bearer <access_token>.
//
// @Summary      Аутентификация (вход)
// @Description  Возвращает access-токен. Неверный логин или пароль — 401. Токен действует JWT_EXPIRES_IN, после logout попадает в блеклист Redis.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body ds.LoginUserRequest true "Логин и пароль"
// @Success      200 {object} ds.TokenResponse
// @Failure      400 {string} string ""
// @Failure      401 {string} string ""
// @Failure      500 {string} string ""
// @Router       /users/auth [post]
func (h *Handler) AuthUser(ctx *gin.Context) {
	var req ds.LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.fail(ctx, http.StatusUnauthorized)
			return
		}
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	if !repository.CheckPassword(user.Password, req.Password) {
		h.fail(ctx, http.StatusUnauthorized)
		return
	}

	token, err := auth.CreateToken(h.JWT, user)
	if err != nil {
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, ds.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.JWT.ExpiresIn.Seconds()),
	})
}

// DeauthUser — POST /api/users/deauth
// Выход: токен из заголовка кладётся в блеклист Redis до истечения exp,
// после чего он перестаёт проходить WithAuthCheck.
//
// @Summary      Выход (logout)
// @Description  Отзывает текущий токен: записывает его в блеклист Redis на срок жизни до exp. Повторное использование токена — 403.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {string} string ""
// @Failure      401 {string} string ""
// @Failure      500 {string} string ""
// @Router       /users/deauth [post]
func (h *Handler) DeauthUser(ctx *gin.Context) {
	raw, ok := bearerToken(ctx)
	if !ok {
		h.fail(ctx, http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseToken(h.JWT, raw)
	if err != nil || claims.ExpiresAt == nil {
		h.fail(ctx, http.StatusUnauthorized)
		return
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		h.fail(ctx, http.StatusUnauthorized)
		return
	}

	if err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), raw, ttl); err != nil {
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
