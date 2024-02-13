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

type RegisterRequestBody struct {
	Username string `json:"username" binding:"required"`
	Room     string `json:"room"`
}

type RollRequestBody struct {
	Dice string `json:"dice" binding:"required"`
}

type DiceResult struct {
	Username string        `json:"username"`
	Room     string        `json:"room"`
	Result   map[byte]byte `json:"result"`
}
