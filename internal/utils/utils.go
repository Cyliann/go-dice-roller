package utils

import (
	"math/rand"

	"github.com/Cyliann/go-dice-roller/internal/stream"
	"github.com/Cyliann/go-dice-roller/internal/types"
)

// RollDice Get number of sides on dice from POST form { "dice": "{"id1": sides, "id2": sides}" } and return the result
func RollDice(username string, room string, dice types.Dice) types.DiceResult {

	roll := (types.Dice)(rand.Intn(int(dice)) + 1)
	diceResult := types.DiceResult{Username: username, Room: room, Result: roll}
	return diceResult
}

func Broadcast(event string, message string, s stream.Stream) {
	msg := types.Message{
		EventType: event,
		Data:      message,
	}
	s.Message <- msg
}
