package server

import (
	//"github.com/dgrijalva/jwt-go"
	//"go/token"
	"io"
	"net/http"

	"github.com/Cyliann/go-dice-roller/internal/utils/token"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

var IDCounter uint = 0

type Greeting struct {
	ID   uint   `json:"id"`
	Name string `json:"username"`
}

type RegistrationInput struct {
	Username string `json:"username" binding:"required"`
}

type RegisteredClient struct {
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

func Register(c *gin.Context) {

	var input RegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newToken, err := token.GenerateToken(uint(10))
	if err != nil {
		log.Error("Error creating JWT for user: ", input.Username)
		return
	}
	client := RegisteredClient{ID: IDCounter, Username: input.Username, Token: newToken}
	c.JSON(http.StatusOK, gin.H{"ID": client.ID, "username": client.Username, "token": client.Token})

	IDCounter++
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
