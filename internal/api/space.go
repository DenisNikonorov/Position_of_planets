package api

import (
	"log"
	"planets/internal/app/handler"
	database "planets/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Server start up")

	db, err := database.NewDB()
	if err != nil {
		logrus.Error("Error initializing database: ", err)
	}
	h := handler.NewHandler(db)

	r := gin.Default()
	r.LoadHTMLGlob("../templates/*")

	r.StaticFile("/style.css", "../style.css")

	r.GET("/feed", h.GetFeed)
	r.GET("/add", h.GetAddPage)
	r.GET("/tile", h.GetTilePage)

	r.Run()

	log.Println("Server down")
}
