package main

import (
	"io"
	"log"
	"net/http"
	"sync"
)

var (
	clients     = make(map[chan<- []byte]struct{})
	clientsLock sync.Mutex
)

func main() {
	http.HandleFunc("/listen", eventsHandler)
	http.HandleFunc("/roll", triggerHandler)
	http.ListenAndServe(":8080", nil)
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	// Log a message when a client connects
	log.Printf("Client connected from %s", r.RemoteAddr)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a channel for this client
	messageChan := make(chan []byte)
	defer close(messageChan)

	// Register the client channel
	clientsLock.Lock()
	clients[messageChan] = struct{}{}
	clientsLock.Unlock()

	// Send a welcome message when a client connects
	welcomeMessage := "Welcome! Connection established.\n\n"
	_, _ = w.Write([]byte(welcomeMessage))
	w.(http.Flusher).Flush()

	// Listen for messages from the channel and send them to the client
	for {
		select {
		case message := <-messageChan:
			_, err := w.Write(message)
			if err != nil {
				// If writing to the response fails, the client might have disconnected.
				log.Printf("Client disconnected from %s. Closing SSE.", r.RemoteAddr)

				// Unregister the client channel
				clientsLock.Lock()
				delete(clients, messageChan)
				clientsLock.Unlock()

				return
			}
			w.(http.Flusher).Flush()
		}
	}
}

func triggerHandler(w http.ResponseWriter, r *http.Request) {
	// Read the POST body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading POST body", http.StatusBadRequest)
		return
	}

	// Broadcast the message to all connected clients
	broadcastMessage(body)

	// Respond to the POST request
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message sent\n"))
}

func broadcastMessage(message []byte) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	// Send the message to all connected clients
	for client := range clients {
		client <- message
	}
}

