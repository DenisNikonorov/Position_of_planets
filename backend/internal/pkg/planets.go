package pkg

import (
	"fmt"

	"planets/internal/app/config"
	"planets/internal/app/handler"
	"planets/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Handler.RegisterHandler(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}

// DefaultMedia — дефолтные фото и видео, которые должны лежать в MinIO.
// Они кладутся в бакет один раз при старте, чтобы ссылка на них
// всегда была рабочей.
func (a *Application) DefaultMedia() map[string]string {
	return map[string]string{
		repository.DefaultPhotoName: a.Config.DefaultPhotoPath,
		repository.DefaultVideoName: a.Config.DefaultVideoPath,
	}
}
