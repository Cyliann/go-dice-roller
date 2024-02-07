package server

import (
	"encoding/json"

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
		var clientID uint = 1
		// Initialize client channel
		name := c.Query("username")
		client := Client{ID: clientID, Chan: make(ClientChan), Name: name}

		// Send new connection to event server
		stream.NewClients <- client

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- client
		}()

		c.Set("client", client)

		// Create a greeting with id
		grt := Greeting{
			ID:   client.ID,
			Name: name,
		}
		clientID++

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
