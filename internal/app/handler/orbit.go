package handler

import (
	"net/http"
	"strconv"

	database "planets/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Database *database.Database
}

func NewHandler(d *database.Database) *Handler {
	return &Handler{
		Database: d,
	}
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	planets, err := h.Database.GetData()
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusInternalServerError, "planets.html", gin.H{"error": err.Error()})
		return
	}

	total := len(planets)
	nextIdx := 0
	if v := ctx.Query("next"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			nextIdx = parsed % total
		}
	}

	ctx.HTML(http.StatusOK, "planets.html", gin.H{
		"planet":       planets[nextIdx],
		"currentIndex": nextIdx,
		"nextIndex":    (nextIdx + 1) % total,
		"total":        total,
	})
}

func (h *Handler) GetAddPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "telescope.html", gin.H{})
}

func (h *Handler) GetTilePage(ctx *gin.Context) {
	maxDist := 1000.0
	if v := ctx.Query("maxDist"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			maxDist = parsed
		}
	}

	planets, err := h.Database.GetPlanetsByMaxDistance(maxDist)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "observatory.html", gin.H{
		"planets": planets,
		"maxDist": maxDist,
	})
}
