package sse

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Cyliann/go-dice-roller/internal/utils/token"
	"github.com/charmbracelet/log"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"net/http"
	"strconv"
)

var IDCounter uint = 0

//type RegistrationInput struct {
//	Username string `json:"username" binding:"required"`
//}

type RequestBody struct {
	Dice string `json:"dice" binding:"required"`
}

type ClientGreeting struct {
	ID       uint
	Room     string
	Username string
	Token    string
}

type DiceResult struct {
	Username string
	Room     string
	Result   int
	//Token    string
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
		if username == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})
			return
		}
		client := Client{ID: IDCounter, Chan: make(ClientChan), Name: username}

		// Send new connection to event server
		stream.NewClients <- client

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- client
		}()

		c.Set("client", client)

		// Register client and create a greeting with id and token
		grt, err := Register(username, roomID)
		if err != nil {
			return
		}

		msg, err := json.Marshal(grt)
		if err != nil {
			log.Error("Error parsing a greeting. Client: %s", c.RemoteIP)
			return
		}

		c.Writer.Write(append(msg, []byte("\n\n")...))
		c.Writer.Flush()
		c.Next()
	}
}

func (s *Server) CreateStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uniuri.NewLen(6)
		s.Streams[id] = NewStream(id)

		c.Request.URL.Path = "/listen/" + id
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

// Handler for post requests that runs RollDice fucntion
func (s *Server) HandleRolls() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Query("username")
		room := c.Param("roomID")
		var requestBody RequestBody

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			log.Error("Error while accessing request body JSON: %s", err)
		}
		//c.JSON(http.StatusOK, requestBody)
		diceResult := RollDice(username, room, requestBody.Dice)
		msg := fmt.Sprintf("User %s rolled %s in the room %s", diceResult.Username, strconv.Itoa(diceResult.Result), diceResult.Room)
		Broadcast("message", msg, s.Streams[room])

	}
}

// Get number of sides on dice from post {"dice": "number" and return the result}
func RollDice(username string, room string, dice string) DiceResult {
	intDice, err := strconv.Atoi(dice)
	if err != nil {
		log.Errorf("Error: can't convert dice number to int %s", dice)
	}
	diceResult := DiceResult{Username: username, Room: room, Result: rand.Intn(intDice) + 1}
	return diceResult
}

func Register(username string, room string) (ClientGreeting, error) {
	newToken, err := token.GenerateToken(IDCounter)
	if err != nil {
		err := "Error creating JWT for user: " + username
		log.Errorf("Error: %s", err)
		return ClientGreeting{0, "", "", ""}, errors.New(err)
	}
	greeting := ClientGreeting{ID: IDCounter, Room: room, Username: username, Token: newToken}
	IDCounter++
	return greeting, nil
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
