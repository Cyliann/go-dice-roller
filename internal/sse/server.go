package sse

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/Cyliann/go-dice-roller/internal/token"
	"github.com/Cyliann/go-dice-roller/internal/types"
	"github.com/charmbracelet/log"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

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

		log.Debugf("New client in room: %s", client.Room)

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
	var requestBody types.RegisterRequestBody

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error while accessing request body JSON": err})
		return
	}
	if requestBody.Username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})
		return
	}

	if requestBody.Room == "" {
		requestBody.Room = s.CreateStream()
	}

	if _, exist := s.Streams[requestBody.Room]; !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Room doesn't exist"})
		return
	}

	newToken, err := token.Generate(requestBody.Username, requestBody.Room)
	if err != nil {
		log.Errorf("Error: %s", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	// Send a cookie
	c.SetSameSite(http.SameSiteDefaultMode)
	c.SetCookie("Authorization", newToken, 3600*24, "", "", false, true)

	c.JSON(http.StatusOK, gin.H{"Room": requestBody.Room})
	log.Debugf("New client registered from %s in room %s", c.ClientIP(), requestBody.Room)
}

// HandleRolls is a handler for post requests that runs RollDice function
func (s *Server) HandleRolls() gin.HandlerFunc {
	return func(c *gin.Context) {

		var requestBody types.RollRequestBody
		var diceArray []uint8

		val, ok := c.Get("client")
		if !ok {
			err := errors.New("Couldn't get client")
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err})
		}
		client := val.(types.Client)

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error while accessing request body JSON": err})
			return
		}

		if err := json.Unmarshal([]byte(requestBody.Dice), &diceArray); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error while parsing JSON to array": err})
			fmt.Printf("Unmarshaled: %v", diceArray)
		}
		fmt.Printf("Unmarshaled array: %v", diceArray)

		diceResult := RollDice(client.Name, client.Room, diceArray)
		msg, err := json.Marshal(diceResult)
		if err != nil {
			log.Errorf("Error parsing dice result: %s", err.Error())
			return
		}
		Broadcast("roll", string(msg), s.Streams[client.Room])

	}
}

func NewServer() Server {
	return Server{Streams: make(map[string]Stream)}
}

// Get number of sides on dice from post { "dice": "number" } and return the result
func RollDice(username string, room string, dice []byte) types.DiceResult {
	var diceArray map[byte]byte
	diceArray = make(map[byte]byte)
	for ind, roll := range dice {
		diceArray[byte(ind)] = byte(rand.Intn(int(roll)) + 1)
	}
	diceResult := types.DiceResult{Username: username, Room: room, Result: diceArray}
	return diceResult
}

func Broadcast(event string, message string, s Stream) {
	msg := types.Message{
		EventType: event,
		Data:      message,
	}
	s.Message <- msg
}
