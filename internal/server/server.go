package server

import (
	"errors"
	"net/http"

	"github.com/Cyliann/go-dice-roller/internal/stream"
	"github.com/Cyliann/go-dice-roller/internal/token"
	"github.com/Cyliann/go-dice-roller/internal/types"
	"github.com/Cyliann/go-dice-roller/internal/utils"
	"github.com/charmbracelet/log"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Addr    string
	Router  *gin.Engine
	Streams map[string]stream.Stream
}

func (s *Server) AddClientToStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, ok := c.Get("client")
		if !ok {
			err := errors.New("couldn't get client")
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
	s.Streams[id] = stream.New(id)
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

	// Send authorization header with the newToken
	c.Header("Access-Control-Allow-Headers", "Authorization")
	c.Header("Authorization", newToken)

	c.JSON(http.StatusOK, gin.H{"Room": requestBody.Room})
	log.Debugf("New client registered from %s in room %s", c.ClientIP(), requestBody.Room)
}

// HandleRolls is a handler for post requests that runs RollDice function
func (s *Server) HandleRolls() gin.HandlerFunc {
	return func(c *gin.Context) {

		var requestBody types.RollRequestBody

		val, ok := c.Get("client")
		if !ok {
			err := errors.New("couldn't get client")
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err})
		}
		client := val.(types.Client)

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error while parsing JSON to array": err})
		}

		diceResult := utils.RollDice(client.Name, client.Room, requestBody.Dice)
		msg, err := diceResult.MarshalJSON()
		if err != nil {
			log.Errorf("Error parsing dice result: %s", err.Error())
			return
		}
		utils.Broadcast("roll", string(msg), s.Streams[client.Room])

	}
}

func New(router *gin.Engine) Server {
	return Server{
		Addr:    ":8080",
		Router:  router,
		Streams: make(map[string]stream.Stream),
	}
}
