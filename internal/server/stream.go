package server

import (
	"encoding/json"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

type Message struct {
	EventType string
	Data      string
}

type Client struct {
	ID   uint
	Chan chan Message
	Name string
}

type ClientChan chan Message

type Stream struct {
	Message       chan Message
	NewClients    chan Client
	ClosedClients chan Client
	TotalClients  map[Client]bool
}

func (stream *Stream) listen() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Infof("Client added. %d registered clients", len(stream.TotalClients))

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client.Chan)
			log.Infof("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Message:
			for client := range stream.TotalClients {
				client.Chan <- eventMsg
			}
		}
	}
}

func (stream *Stream) ServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		username := c.Query("username")
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})
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
		grt := Register(username)
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

func NewStream() Stream {
	stream := Stream{
		Message:       make(chan Message),
		NewClients:    make(chan Client),
		ClosedClients: make(chan Client),
		TotalClients:  make(map[Client]bool),
	}

	go stream.listen()

	return stream
}
