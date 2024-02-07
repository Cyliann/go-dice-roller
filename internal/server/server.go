package server

import (
	"io"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

var ID uint32 = 0

type Greeting struct {
	ID   uint32 `json:"id"`
	Name string `json:"username"`
}

func HandleClients(c *gin.Context) {
	v, ok := c.Get("client")
	if !ok {
		log.Warn("Couldn't get client")
		return
	}
	client, ok := v.(Client)
	if !ok {
		return
	}
	c.Stream(func(w io.Writer) bool {
		// Stream message to client from message channel
		if msg, ok := <-client.Chan; ok {
			c.SSEvent(msg.EventType, msg.Data)
			return true
		}
		return false
	})
}

func Broadcast(event string, message string, stream Stream) {
	msg := Message{
		EventType: event,
		Data:      message,
	}
	stream.Message <- msg
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}
