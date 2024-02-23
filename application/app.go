package application

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

type App struct {
	router *gin.Engine
}

func New() App {
	return App{
		router: loadRoutes(),
	}
}

func (a *App) Start() error {
	log.SetLevel(log.DebugLevel)
	server := &http.Server{
		Addr:    ":8080",
		Handler: a.router,
	}

	log.Info("Listening on port 8080...")
	err := server.ListenAndServe()

	if err != nil {
		return fmt.Errorf("Failed to start server: %w", err)
	}

	return nil
}
