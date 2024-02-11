package main

import (
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	"github.com/Cyliann/go-dice-roller/internal/middleware"
	"github.com/Cyliann/go-dice-roller/internal/sse"
	"github.com/Cyliann/go-dice-roller/internal/token"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	log.SetLevel(log.DebugLevel)
	router := gin.Default()
	s := sse.NewServer()

	router.GET("/play", token.Validate(), middleware.Headers(), s.AddClientToStream(), middleware.HandleClients())
	router.POST("/register", s.Register)
	// POST form: { "dice" : "[number of sides]" }
	router.POST("/roll", token.Validate(), s.HandleRolls())

	log.Info("Listening on 8080...")
	router.Run(":8080")
}
