package main

import (
	"github.com/Cyliann/go-dice-roller/application"
	"github.com/charmbracelet/log"
)

func main() {
	app := application.New()

	err := app.Start()

	if err != nil {
		log.Errorf("Failer to start app: %s", err)
	}

}
