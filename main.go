package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"

	"github.com/charmbracelet/log"
)

type Greeting struct {
	ID uint32 `json:"id"`
}

type RollRequestPayload struct {
	ID   uint32 `json:"id"`
	Dice uint8  `json:"dice"`
}

type RollResponsePayload struct {
	ID     uint32 `json:"id"`
	Result uint8  `json:"result"`
}

var (
	clients     = make(map[chan<- []byte]struct{})
	clientsLock sync.Mutex
	id_counter  uint32 = 0
	port               = 8080
)

func main() {
	log.SetLevel(log.DebugLevel)
	http.HandleFunc("/listen", clientHandler)
	http.HandleFunc("/roll", triggerHandler)

	fmt.Printf("Welcome to Roll Dicer!\nListening on port %d...\n\n", port)
	log.Error(http.ListenAndServe(":"+fmt.Sprint(port), nil))
	fmt.Print("Exiting...")
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	// Log a message when a client connects
	log.Debugf("Client connected from %s", r.RemoteAddr)

	// Send a welcome message when a client connects
	greet(w, r)

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

	// Listen for messages from the channel and send them to clients
	for {
		select {
		case message := <-messageChan:
			_, err := w.Write(message)
			if err != nil {
				// If writing to the response fails, the client might have disconnected.
				log.Infof("Client disconnected from %s. Closing SSE.", r.RemoteAddr)

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

func greet(w http.ResponseWriter, r *http.Request) {
	var greeting Greeting
	greeting.ID = id_counter
	id_counter++

	greetingMessage, err := json.Marshal(greeting)

	if err != nil {
		log.Fatalf("Error assigning ID: %d to a client %s", id_counter, r.RemoteAddr)
	}

	w.Write([]byte(greetingMessage))
	w.(http.Flusher).Flush()
}

func triggerHandler(w http.ResponseWriter, r *http.Request) {
	// Read the POST body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading POST body", http.StatusBadRequest)
		return
	}

	var requestPayload RollRequestPayload
	if err := json.Unmarshal(body, &requestPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	var responsePyaload RollResponsePayload
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
		log.Error("Error encoding struct to JSON", "err", err)
		return
	}

	// Send the message to all connected clients
	for client := range clients {
		client <- jsonMessage
	}
	log.Debugf("A message has been broadcasted: %s", jsonMessage)
}
