package main

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ThrowRequest struct {
	ID       int `json:"id"`
	DiceType int `json:"dice"`
}

type ThrowResult struct {
	ID     int `json:"id"`
	Result int `json:"result"`
}

func main() {
	router := gin.Default()
	router.POST("/throw", PostThrow)

	router.Run("localhost:8080")
}

func PostThrow(c *gin.Context) {
	var newThrowRequest ThrowRequest
	var newThrowResult ThrowResult

	if err := c.BindJSON(&newThrowRequest); err != nil {
		return
	}

	newThrowResult.ID = newThrowRequest.ID
	newThrowResult.Result = rand.Intn(newThrowRequest.DiceType) + 1

	c.JSON(http.StatusCreated, newThrowResult)
}
