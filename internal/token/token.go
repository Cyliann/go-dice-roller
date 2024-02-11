package token

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Cyliann/go-dice-roller/internal/types"
	"github.com/charmbracelet/log"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var (
	IDCounter     uint = 0
	signingSecret      = uniuri.New() // generate random secret
)

func Generate(username string, room string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      IDCounter,
		"room":     room,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // expires in 1 day
	})
	IDCounter++

	return token.SignedString([]byte(signingSecret))
}

func Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("Authorization")

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// From: https://pkg.go.dev/github.com/golang-jwt/jwt@v3.2.2+incompatible#Parse
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(signingSecret), nil
		})

		if err != nil {
			log.Errorf(err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Errorf("Invalid claims or token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Invalid claims or token"})
			return
		}

		client := types.Client{
			ID:   uint(claims["sub"].(float64)),
			Name: claims["username"].(string),
			Room: claims["room"].(string),
			Chan: make(types.ClientChan),
		}

		c.Set("client", client)
		c.Next()
	}
}
