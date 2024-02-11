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

type RequestBody struct {
	Dice uint8 `json:"dice" binding:"required"`
}

type DiceResult struct {
	Username string
	Room     string
	Result   int
}
