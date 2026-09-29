package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"

	"planets/internal/app/auth"
	"planets/internal/app/ds"
	"planets/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler регистрирует маршруты веб-сервиса. Все адреса
// начинаются с /api, домен услуг — /planets.
func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.GET("/planets", h.GetAllPlanets)
		api.POST("/planets", h.AddPlanet)
		api.GET("/planets/draft", h.GetDraft)
		api.PUT("/planets/:id", h.PublishPlanet)
		api.GET("/planets/feed", h.GetFeed)
		api.POST("/planets/:id/like", h.LikePlanet)
		api.DELETE("/planets/:id", h.DeletePlanet)

		api.POST("/users", h.RegisterUser)
		api.POST("/users/auth", h.AuthUser)
		api.POST("/users/deauth", h.DeauthUser)
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
func (h *Handler) GetAllPlanets(ctx *gin.Context) {
	var maxDist *float64
	if raw := ctx.Query("maxDist"); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			maxDist = &parsed
		}
	}

	planets, err := h.Repository.GetPlanets(ctx.Query("search"), maxDist, auth.CurrentUserID())
	if err != nil {
		h.fail(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, planets)
}

// GetFeed — GET /api/planets/feed?id=&next=true
// Лента опубликованных планет, листание по кругу.
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

	planet, err := h.Repository.GetFeed(id, ctx.Query("next") == "true", auth.CurrentUserID())
	if err != nil {
		h.fail(ctx, http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, planet)
}

// GetDraft — GET /api/planets/draft
// Черновик текущего пользователя, идентификатор не передаётся.
func (h *Handler) GetDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetUserDraft(auth.CurrentUserID())
	if err != nil {
		h.fail(ctx, http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

// AddPlanet — POST /api/planets
// multipart/form-data: name, description, distance_from_sun, perihelion,
// observation_date, а также файлы под ключами image_url и url_video.
// Создаёт запись в статусе draft и складывает медиафайлы в MinIO.
// Возвращает 201 и заполненный JSON.
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

	planet, err := h.Repository.CreatePlanet(auth.CurrentUserID(), req, imageName, videoName)
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

// PublishPlanet — PUT /api/planets/:id
// Публикация черновика. Тело не требуется: бизнес-поля уже заполнены
// при создании, здесь меняется только статус.
func (h *Handler) PublishPlanet(ctx *gin.Context) {
	id, ok := planetID(ctx)
	if !ok {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	planet, err := h.Repository.PublishPlanet(auth.CurrentUserID(), id)
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

	planet, err := h.Repository.SetLike(auth.CurrentUserID(), id, *req.Like)
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
func (h *Handler) DeletePlanet(ctx *gin.Context) {
	id, ok := planetID(ctx)
	if !ok {
		h.fail(ctx, http.StatusBadRequest)
		return
	}

	planet, err := h.Repository.SoftDeletePlanet(auth.CurrentUserID(), id)
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

// RegisterUser — POST /api/users
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
// Заглушка: реальная аутентификация появится в ЛР4.
func (h *Handler) AuthUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{})
}

// DeauthUser — POST /api/users/deauth
// Заглушка: реальная деавторизация появится в ЛР4.
func (h *Handler) DeauthUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{})
}
