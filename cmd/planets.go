// Package main — точка входа веб-сервиса «Положение планет».
//
// @title                        Planets Position API
// @version                      1.0
// @description                  Бэкенд SPA с планетами. Аутентификация по JWT: access-токен выдаёт POST /api/users/auth и проверяется middleware, logout кладёт токен в блеклист Redis.
// @BasePath                     /api
// @schemes                      http
// @securityDefinitions.apikey   BearerAuth
// @in                           header
// @name                         Authorization
// @description                  Формат: "Bearer <access_token>" из ответа POST /api/users/auth. Без токена доступно только чтение.
package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"planets/internal/app/config"
	"planets/internal/app/dsn"
	"planets/internal/app/handler"
	"planets/internal/app/redis"
	"planets/internal/app/repository"
	"planets/internal/pkg"

	_ "planets/docs"
)

func main() {
	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	redisClient, err := redis.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error connecting to redis: %v", err)
	}
	defer redisClient.Close()

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

	hand := handler.NewHandler(rep, redisClient, conf.JWT)

	application := pkg.NewApp(conf, router, hand)

	if err := rep.EnsureDefaults(application.DefaultMedia()); err != nil {
		logrus.Fatalf("error uploading default media: %v", err)
	}
	logrus.Info("default media ready")

	application.RunApp()
}
