package application

import (
	"github.com/Cyliann/go-dice-roller/internal/middleware"
	"github.com/Cyliann/go-dice-roller/internal/server"
	"github.com/Cyliann/go-dice-roller/internal/token"
	"github.com/gin-gonic/gin"
)

func loadRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	s := server.New(router)

	router.GET("/play", token.Validate(), middleware.Headers(), s.AddClientToStream(), middleware.HandleClients())
	router.POST("/register", s.Register)
	// POST form: { "dice" : "[number of sides]" }
	router.POST("/roll", token.Validate(), s.HandleRolls())

	return router
}
