package utils

import (
	"math/rand"

	"github.com/Cyliann/go-dice-roller/internal/stream"
	"github.com/Cyliann/go-dice-roller/internal/types"
)

// RollDice Get number of sides on dice from POST form { "dice": "{"id1": sides, "id2": sides}" } and return the result
func RollDice(username string, room string, dice types.DiceArray) types.DiceResult {

	var diceArray types.DiceArray

	// The id is kept from the original POST form, and there is a random roll assigned to it
	for _, diceSides := range dice {
		diceArray = append(diceArray, uint8(rand.Intn(int(diceSides))+1))
	}
	diceResult := types.DiceResult{Username: username, Room: room, Result: diceArray}
	return diceResult
}

func Broadcast(event string, message string, s stream.Stream) {
	msg := types.Message{
		EventType: event,
		Data:      message,
	}
	s.Message <- msg
}
