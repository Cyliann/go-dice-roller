package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// for now only username, later -> password
type RegistrationInput struct {
	Username string `json:"username" binding:"required"`
}

func Register(c *gin.Context) {

	var input RegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "validated!"})
}
