package main

import (
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	"github.com/Cyliann/go-dice-roller/internal/sse"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	s := sse.NewServer()

	router.GET("/play", sse.HeadersMiddleware(), s.AddClientToStream(), sse.HandleClients())
	router.GET("/register/:roomID", s.Register)
	// POST form: { "dice" : "[number of sides]" }
	router.POST("/roll/:roomID", s.HandleRolls())

	log.Info("Listening on 8080...")
	router.Run(":8080")
}
