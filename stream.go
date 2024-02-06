package main

import (
	"encoding/json"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

type Message struct {
	EventType string
	Data      string
}

type ClientChan chan Message

type Stream struct {
	Message       chan Message
	NewClients    chan chan Message
	ClosedClients chan chan Message
	TotalClients  map[chan Message]bool
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
			close(client)
			log.Infof("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Message:
			for clientMessageChan := range stream.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}

func (stream *Stream) ServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		stream.NewClients <- clientChan

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- clientChan
		}()

		c.Set("clientChan", clientChan)

		// Create a greeting with id
		grt := Greeting{
			ID: ID,
		}
		ID++
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

func New() Stream {
	stream := Stream{
		Message:       make(chan Message),
		NewClients:    make(chan chan Message),
		ClosedClients: make(chan chan Message),
		TotalClients:  make(map[chan Message]bool),
	}

	go stream.listen()

	return stream
}
