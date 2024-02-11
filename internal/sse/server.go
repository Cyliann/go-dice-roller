package sse

import (
	"encoding/json"
	"github.com/Cyliann/go-dice-roller/internal/utils/token"
	"github.com/charmbracelet/log"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"net/http"
)

type RequestBody struct {
	Dice uint8 `json:"dice" binding:"required"`
}

type ClientGreeting struct {
	ID       uint   `json:"id"`
	Room     string `json:"room"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type DiceResult struct {
	Username string
	Room     string
	Result   int
}

type Server struct {
	Streams map[string]Stream
}

func (s *Server) AddClientToStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("roomID")

		stream, exist := s.Streams[roomID]
		if !exist {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Room doesn't exist"})
			return
		}

		// Initialize client channel
		username := c.Query("username")
		client := Client{ID: 0, Chan: make(ClientChan), Name: username}

		// Send new connection to event server
		stream.NewClients <- client

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- client
		}()

		c.Set("client", client)

		c.Next()
	}
}

func (s *Server) CreateStream() string {
	id := uniuri.NewLen(6)
	s.Streams[id] = NewStream(id)

	return id
}

// Handler for post requests that runs RollDice fucntion
func (s *Server) HandleRolls() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Query("username")
		room := c.Param("roomID")
		var requestBody RequestBody

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error while accessing request body JSON": err})
		}
		diceResult := RollDice(username, room, requestBody.Dice)
		msg, err := json.Marshal(diceResult)
		if err != nil {
			log.Errorf("Error parsing dice result", err)
			return
		}
		Broadcast("roll", string(msg), s.Streams[room])

	}
}

func NewServer() Server {
	return Server{Streams: make(map[string]Stream)}
}

func HandleClients() gin.HandlerFunc {
	return func(c *gin.Context) {
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
}

// Get number of sides on dice from post {"dice": "number" and return the result}
func RollDice(username string, room string, dice uint8) DiceResult {
	diceResult := DiceResult{Username: username, Room: room, Result: rand.Intn(int(dice)) + 1}
	return diceResult
}

func (s *Server) Register(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})
		return
	}

	room := c.Param("roomID")
	if room == "" {
		room = s.CreateStream()
	}
	newToken, err := token.GenerateToken(username, room)
	if err != nil {
		// err := "Error creating JWT for user: " + username
		log.Errorf("Error: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
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
