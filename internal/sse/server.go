package sse

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"

	"github.com/Cyliann/go-dice-roller/internal/utils/token"
)

var IDCounter uint = 0

type RegistrationInput struct {
	Username string `json:"username" binding:"required"`
}

type ClientGreeting struct {
	ID       uint
	Room     string
	Username string
	Token    string
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

func Register(username string, room string) (ClientGreeting, error) {
	newToken, err := token.GenerateToken(uint(10))
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
