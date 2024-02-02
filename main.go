package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

type rollRequestPayload struct {
	ID   uint32 `json:"id"`
	Dice uint8  `json:"dice"`
}

type rollResponsePayload struct {
	ID     uint32 `json:"id"`
	Result uint8  `json:"result"`
}

var (
	clients     = make(map[chan<- []byte]struct{})
	clientsLock sync.Mutex
)

func main() {
	var port = 8080
	http.HandleFunc("/listen", eventsHandler)
	http.HandleFunc("/roll", triggerHandler)

	fmt.Printf("Welcome to Roll Dicer!\nListening on port %d...\n", port)
	http.ListenAndServe(":"+fmt.Sprint(port), nil)
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

	var requestPayload rollRequestPayload
	if err := json.Unmarshal(body, &requestPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	var responsePyaload rollResponsePayload
	responsePyaload.ID = requestPayload.ID
	responsePyaload.Result = uint8(rand.Intn(int(requestPayload.Dice)) + 1)

	// Broadcast the message to all connected clients
	broadcastMessage(responsePyaload)

	// Respond to the POST request
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message sent\n"))
}

func broadcastMessage(message interface{}) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	// Convert the struct to JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Println("Error encoding struct to JSON:", err)
		return
	}

	// Send the message to all connected clients
	for client := range clients {
		client <- jsonMessage
	}
}
