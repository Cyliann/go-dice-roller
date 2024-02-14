package types

import (
	"fmt"
	"strings"
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

type DiceArray []uint8

type RollRequestBody struct {
	Dice DiceArray `json:"dice" binding:"required"`
}

type DiceResult struct {
	Username string    `json:"username"`
	Room     string    `json:"room"`
	Result   DiceArray `json:"result"`
}

// From: https://stackoverflow.com/questions/14177862/how-to-marshal-a-byte-uint8-array-as-json-array-in-go
func (dr *DiceResult) MarshalJSON() ([]byte, error) {
	var array string
	if dr.Result == nil {
		array = "null"
	} else {
		array = strings.Join(strings.Fields(fmt.Sprintf("%d", dr.Result)), ",")
	}
	jsonResult := fmt.Sprintf(`{"Username":%q, "Room":%q, "Result":%s}`, dr.Username, dr.Room, array)
	return []byte(jsonResult), nil
}
