package types

type Message struct {
	EventType string
	Data      string
}

type Client struct {
	ID   uint
	Name string
	Room string
	Chan chan Message
}

type ClientChan chan Message
