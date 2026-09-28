package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"planets/internal/app/ds"
	"planets/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.GetAllPlanets)
	router.GET("/planets", h.GetAllPlanets)
	router.GET("/planets/:id", h.GetPlanetById)
	router.POST("/delete-planet", h.DeletePlanet)

	router.GET("/feed", h.GetFeed)
	router.GET("/add", h.GetAddPage)
	router.POST("/add", h.PostAddPage)
	router.GET("/add/:id", h.GetAddStep2)
	router.POST("/add/:id", h.PostAddStep2)
	router.GET("/tile", h.GetTilePage)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
	router.StaticFile("/media/default_video.mp4", "resources/default_video.mp4")
	router.StaticFile("/media/default_photo.png", "resources/default_photo.png")
	router.GET("/style.css", func(ctx *gin.Context) {
		// Не даём браузерам кэшировать старый CSS: no-cache заставляет
		// перезапросить файл при каждом изменении.
		ctx.Header("Cache-Control", "no-cache")
		ctx.File("style.css")
	})
}

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	planets, err := h.Repository.GetAllPlanets()
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusInternalServerError, "planets.html", gin.H{"error": err.Error()})
		return
	}

	total := len(planets)
	emptyPlanet := ds.Planet{}
	nextIdx := 0
	if v := ctx.Query("id"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			for i, p := range planets {
				if p.ID == parsed {
					nextIdx = i
					break
				}
			}
		}
	} else if v := ctx.Query("next"); v != "" && total > 0 {
		if parsed, err := strconv.Atoi(v); err == nil {
			nextIdx = parsed % total
		}
	}

	planet := emptyPlanet
	if total > 0 {
		planet = planets[nextIdx%total]
	}

	nextIndex := 0
	if total > 0 {
		nextIndex = (nextIdx + 1) % total
	}

	ctx.HTML(http.StatusOK, "planets.html", gin.H{
		"planet":       planet,
		"currentIndex": nextIdx,
		"nextIndex":    nextIndex,
		"total":        total,
	})
}

func (h *Handler) GetAddPage(ctx *gin.Context) {
	draft, err := h.Repository.GetFirstDraft()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draft != nil {
		ctx.Redirect(http.StatusFound, fmt.Sprintf("/add/%d", draft.ID))
		return
	}

	ctx.HTML(http.StatusOK, "telescope.html", gin.H{})
}

// PostAddPage обрабатывает шаг 1: загрузка медиа + название → черновик планеты.
// Если карточка с таким названием уже существует — переходим на шаг 2 с её данными.
func (h *Handler) PostAddPage(ctx *gin.Context) {
	name := ctx.PostForm("name")
	if name == "" {
		ctx.Redirect(http.StatusFound, "/add")
		return
	}

	existing, err := h.Repository.FindPlanetByName(name)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if existing != nil {
		ctx.Redirect(http.StatusFound, fmt.Sprintf("/add/%d", existing.ID))
		return
	}

	var imageURL string
	var urlVideo string
	form, err := ctx.MultipartForm()
	if err == nil {
		if files, ok := form.File["media"]; ok && len(files) > 0 {
			objectName, hasVideo, err := h.Repository.UploadImage(files[0])
			if err != nil {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
				return
			}
			if hasVideo {
				urlVideo = objectName
			} else {
				imageURL = objectName
			}
		}
	}

	planetID, err := h.Repository.CreatePlanetDraft(name, imageURL, urlVideo)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, fmt.Sprintf("/add/%d", planetID))
}

// GetAddStep2 показывает шаг 2: превью и поля для публикации черновика.
func (h *Handler) GetAddStep2(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Redirect(http.StatusFound, "/add")
		return
	}

	planet, err := h.Repository.GetPlanetForEdit(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if planet == nil {
		ctx.Redirect(http.StatusFound, "/planets")
		return
	}

	ctx.HTML(http.StatusOK, "planet_publish.tmpl", gin.H{
		"planet": planet,
	})
}

// PostAddStep2 публикует черновик (шаг 2) и ведёт на страницу планеты.
func (h *Handler) PostAddStep2(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Redirect(http.StatusFound, "/add")
		return
	}

	description := ctx.PostForm("description")
	distance, err1 := strconv.ParseFloat(ctx.PostForm("distance"), 64)
	observationDate := ctx.PostForm("observation_date")
	if err1 != nil {
		ctx.Redirect(http.StatusFound, fmt.Sprintf("/add/%d", id))
		return
	}

	err = h.Repository.PublishPlanet(id, description, distance, observationDate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, fmt.Sprintf("/feed?id=%d", id))
}

func (h *Handler) GetTilePage(ctx *gin.Context) {
	maxDist := 1000.0
	if v := ctx.Query("maxDist"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			maxDist = parsed
		}
	}

	planets, err := h.Repository.GetPlanetsByMaxDistance(maxDist)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "observatory.html", gin.H{
		"planets": planets,
		"maxDist": maxDist,
	})
}

func (h *Handler) GetAllPlanets(ctx *gin.Context) {
	var planets []ds.Planet
	var err error

	search := ctx.Query("search")
	if search == "" {
		planets, err = h.Repository.GetAllPlanets()
	} else {
		planets, err = h.Repository.SearchPlanetsByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "planets_list.tmpl", gin.H{
		"data":        planets,
		"liked_count": h.Repository.GetUserLikesCount(1),
		"search":      search,
	})
}

func (h *Handler) GetPlanetById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	planet, err := h.Repository.GetPlanetByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}
	if planet == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "планета не найдена",
		})
		logrus.Errorf("planet with id %d not found", id)
		return
	}

	ctx.HTML(http.StatusOK, "planet_detail.tmpl", gin.H{
		"planet": planet,
	})
}

func (h *Handler) DeletePlanet(ctx *gin.Context) {
	strId := ctx.PostForm("planet_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = h.Repository.DeletePlanet(uint(id))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	redirectTo := ctx.PostForm("redirect_to")
	if redirectTo == "" {
		redirectTo = "/planets"
	}
	ctx.Redirect(http.StatusFound, redirectTo)
}
