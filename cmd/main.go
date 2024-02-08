package main

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	"github.com/Cyliann/go-dice-roller/internal/server"
)

// type RollRequestPayload struct {
// 	ID   uint32 `json:"id"`
// 	Dice uint8  `json:"dice"`
// }
//
// type RollResponsePayload struct {
// 	ID     uint32 `json:"id"`
// 	Result uint8  `json:"result"`
// }

func main() {
	router := gin.Default()
	stream := server.NewStream()

	router.GET("/listen", server.HeadersMiddleware(), (&stream).ServeHTTP(), server.HandleClients)

	go func() {
		for {
			server.Broadcast("message", "It works!", stream)
			time.Sleep(time.Second * 2)
		}
	}()
	log.Info("Listening on 8080...")
	router.Run(":8080")
}
