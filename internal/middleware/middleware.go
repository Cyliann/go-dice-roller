package middleware

import (
	"io"

	"github.com/Cyliann/go-dice-roller/internal/types"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

func HandleClients() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("client")
		if !ok {
			log.Warn("Couldn't get client")
			return
		}
		client, ok := v.(types.Client)
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
}

func Headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}
