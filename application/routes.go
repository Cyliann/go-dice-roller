package application

import (
	"time"

	_ "github.com/Cyliann/go-dice-roller/docs"
	"github.com/Cyliann/go-dice-roller/internal/middleware"
	"github.com/Cyliann/go-dice-roller/internal/server"
	"github.com/Cyliann/go-dice-roller/internal/token"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func loadRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://diceroll.cych.eu"}, // Frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		MaxAge:           time.Duration(time.Hour),
		AllowCredentials: true,
	}))
	s := server.New(router)

	router.GET("/play", token.Validate(), middleware.Headers(), s.AddClientToStream(), middleware.HandleClients())
	router.POST("/register", s.Register)
	// POST form: { "dice" : "number of sides" }
	router.POST("/roll", token.Validate(), s.HandleRolls())

	// Docs
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
