package api

import (
	"fmt"
	"net/http"

	"github.com/Koubae/GoAnyBusiness/internal/app/core"
	"github.com/gin-gonic/gin"
)

type IndexController struct{}

func (controller *IndexController) Index(c *gin.Context) {
	config := core.GetDefaultConfig()

	response := []byte(fmt.Sprintf("Welcome to %s V%s", config.AppName, config.AppVersion))
	c.Data(http.StatusOK, "text/html; charset=utf-8", response)

}

func (controller *IndexController) Ping(c *gin.Context) {
	response := []byte("pong")
	c.Data(http.StatusOK, "text/html; charset=utf-8", response)
}

func (controller *IndexController) Alive(c *gin.Context) {
	response := []byte("OK")
	c.Data(http.StatusOK, "text/html; charset=utf-8", response)
}

func (controller *IndexController) Ready(c *gin.Context) {
	response := []byte("OK")
	c.Data(http.StatusOK, "text/html; charset=utf-8", response)
}
