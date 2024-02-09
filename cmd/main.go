package main

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	"github.com/Cyliann/go-dice-roller/internal/sse"
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	s := sse.NewServer()

	router.GET("/listen/:roomID", sse.HeadersMiddleware(), s.AddClientToStream(), sse.HandleClients())
	router.GET("/listen", s.CreateStream(), func(c *gin.Context) { router.HandleContext(c) })
	// POST form: { "dice" : "[number of sides]" }
	router.POST("/listen/:roomID", s.HandleRolls())

	// Loop through all streams and send a test message
	go func() {
		for {
			for _, stream := range s.Streams {
				sse.Broadcast("message", "It works!", stream)
			}
			time.Sleep(time.Second * 2)
		}
	}()
	log.Info("Listening on 8080...")
	router.Run(":8080")
}
