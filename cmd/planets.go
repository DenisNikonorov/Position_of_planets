package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"planets/internal/app/config"
	"planets/internal/app/dsn"
	"planets/internal/app/handler"
	"planets/internal/app/repository"
	"planets/internal/pkg"
)

func main() {
	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	rep, err := repository.New(repository.RepositorySettings{
		PostgresDSN:    dsn.FromEnv(),
		MinioEndpoint:  conf.MinioEndpoint,
		MinioAccessKey: conf.MinioAccessKey,
		MinioSecretKey: conf.MinioSecretKey,
		MinioBucket:    conf.MinioBucket,
		MinioUseSSL:    conf.MinioUseSSL,
		MinioPublicURL: conf.MinioPublicURL,
	})
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)

	if err := rep.EnsureDefaults(application.DefaultMedia()); err != nil {
		logrus.Fatalf("error uploading default media: %v", err)
	}
	logrus.Info("default media ready")

	application.RunApp()
}
