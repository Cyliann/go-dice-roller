package application

import (
	_ "github.com/Cyliann/go-dice-roller/docs"
	"github.com/Cyliann/go-dice-roller/internal/middleware"
	"github.com/Cyliann/go-dice-roller/internal/server"
	"github.com/Cyliann/go-dice-roller/internal/token"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func loadRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	s := server.New(router)

	router.GET("/play", token.Validate(), middleware.Headers(), s.AddClientToStream(), middleware.HandleClients())
	router.POST("/register", s.Register)
	// POST form: { "dice" : "number of sides" }
	router.POST("/roll", token.Validate(), s.HandleRolls())

	// Docs
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
