package sse

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"

	"github.com/Cyliann/go-dice-roller/internal/token"
	"github.com/Cyliann/go-dice-roller/internal/types"
	"github.com/charmbracelet/log"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

type RequestBody struct {
	Dice uint8 `json:"dice" binding:"required"`
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
		val, ok := c.Get("client")
		if !ok {
			err := errors.New("Couldn't get client")
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err})
		}
		client := val.(types.Client)

		log.Infof("Client room: %s", client.Room)

		stream, exist := s.Streams[client.Room]
		if !exist {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Room doesn't exist"})
			return
		}

		// Send new connection to event server
		stream.NewClients <- client

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- client
		}()

		c.Next()
	}
}

func (s *Server) CreateStream() string {
	id := uniuri.NewLen(6)
	s.Streams[id] = NewStream(id)
	log.Infof("Room: %s", id)

	return id
}

func (s *Server) Register(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})
		return
	}

	room := c.Param("roomID")
	log.Infof("roomID: %s", room)
	if room == "/" {
		room = s.CreateStream()
	}
	newToken, err := token.Generate(username, room)
	if err != nil {
		// err := "Error creating JWT for user: " + username
		log.Errorf("Error: %s", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	// Send a cookie
	c.SetSameSite(http.SameSiteDefaultMode)
	c.SetCookie("Authorization", newToken, 3600*24, "", "", false, true)

	c.JSON(http.StatusOK, gin.H{})
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
			log.Errorf("Error parsing dice result: %s", err.Error())
			return
		}
		Broadcast("roll", string(msg), s.Streams[room])

	}
}

func NewServer() Server {
	return Server{Streams: make(map[string]Stream)}
}

// Get number of sides on dice from post {"dice": "number" and return the result}
func RollDice(username string, room string, dice uint8) DiceResult {
	diceResult := DiceResult{Username: username, Room: room, Result: rand.Intn(int(dice)) + 1}
	return diceResult
}

func Broadcast(event string, message string, s Stream) {
	msg := types.Message{
		EventType: event,
		Data:      message,
	}
	s.Message <- msg
}
