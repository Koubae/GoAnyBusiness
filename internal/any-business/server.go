package any_business

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %s", err.Error()))
	}
}

func Run() {
	router := gin.Default()
	router.GET(
		"/ping", func(c *gin.Context) {
			c.JSON(
				200, gin.H{
					"message": "pong",
				},
			)
		},
	)
	router.Run() // listens on 0.0.0.0:8080 by default
}
