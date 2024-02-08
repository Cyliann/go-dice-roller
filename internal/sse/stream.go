package sse

import (
	"github.com/charmbracelet/log"
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

func NewStream(id string) Stream {
	stream := Stream{
		Message:       make(chan Message),
		NewClients:    make(chan Client),
		ClosedClients: make(chan Client),
		TotalClients:  make(map[Client]bool),
	}

	go stream.listen()

	return stream
}