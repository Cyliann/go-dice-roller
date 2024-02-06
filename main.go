package main

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

// type Greeting struct {
// 	ID uint32 `json:"id"`
// }
//
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
	stream := New()

	router.GET("/listen", HeadersMiddleware(), (&stream).ServeHTTP(), handleClients)

	go func() {
		for {
			broadcast("message", "It works!", stream)
			time.Sleep(time.Second * 2)
		}
	}()
	log.Info("Listening on 8080...")
	router.Run(":8080")
}
