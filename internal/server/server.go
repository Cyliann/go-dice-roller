package server

import (
	"io"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	"github.com/Cyliann/go-dice-roller/internal/utils/token"
)

var IDCounter uint = 0

type Greeting struct {
	ID   uint   `json:"id"`
	Name string `json:"username"`
}

type RegistrationInput struct {
	Username string `json:"username" binding:"required"`
}

type ClientGreeting struct {
	ID       uint
	Username string
	Token    string
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

func Register(username string) ClientGreeting {
	newToken, err := token.GenerateToken(uint(10))
	if err != nil {
		log.Error("Error creating JWT for user: ", username)
		return ClientGreeting{0, "", ""}
	}
	greeting := ClientGreeting{ID: IDCounter, Username: username, Token: newToken}
	IDCounter++
	return greeting
}

func Broadcast(event string, message string, s Stream) {
	msg := Message{
		EventType: event,
		Data:      message,
	}
	s.Message <- msg
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
