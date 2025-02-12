package types

import (
	"fmt"
)

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

type Dice uint8

type RollRequestBody struct {
	Dice Dice `json:"dice" binding:"required"`
}

type DiceResult struct {
	Username string `json:"username"`
	Room     string `json:"room"`
	Result   Dice   `json:"result"`
}

// From: https://stackoverflow.com/questions/14177862/how-to-marshal-a-byte-uint8-array-as-json-array-in-go
func (dr *DiceResult) MarshalJSON() ([]byte, error) {
	jsonResult := fmt.Sprintf(`{"Username":%q, "Room":%q, "Result":%d}`, dr.Username, dr.Room, dr.Result)
	return []byte(jsonResult), nil
}
