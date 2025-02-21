package stream

import (
	"github.com/Cyliann/go-dice-roller/internal/types"
	"github.com/charmbracelet/log"
)

type Stream struct {
	Message       chan types.Message
	NewClients    chan types.Client
	ClosedClients chan types.Client
	TotalClients  map[types.Client]bool
}

func (stream *Stream) listen() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Infof("Client added. %d registered clients", len(stream.TotalClients))

			eventMsg := types.Message{EventType: "join", Data: client.Name}
			stream.send_message(eventMsg)

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client.Chan)
			log.Infof("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Message:
			stream.send_message(eventMsg)
		}
	}
}

func (stream *Stream) send_message(eventMsg types.Message) {
	for client := range stream.TotalClients {
		client.Chan <- eventMsg
	}
}

func New(id string) Stream {
	stream := Stream{
		Message:       make(chan types.Message),
		NewClients:    make(chan types.Client),
		ClosedClients: make(chan types.Client),
		TotalClients:  make(map[types.Client]bool),
	}

	go stream.listen()

	return stream
}
