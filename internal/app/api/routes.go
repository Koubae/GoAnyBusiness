package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Koubae/GoAnyBusiness/internal/app/core"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CreateRouter creates a new router
func CreateRouter(config *core.Config) *http.Handler {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.Use(
		cors.New(
			cors.Config{
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
				ExposeHeaders:    []string{"Content-Length"},
				MaxAge:           12 * time.Hour,
				AllowCredentials: false,
				AllowAllOrigins:  config.Env != core.Production,
			},
		),
	)
	err := router.SetTrustedProxies(config.TrustedProxies)
	if err != nil {
		log.Fatalf("Error setting trusted proxies, error: %s", err.Error())
	}

	index := router.Group("/")
	{
		index.GET(
			"/", func(c *gin.Context) {
				c.Data(
					http.StatusOK,
					"text/html; charset=utf-8",
					[]byte(fmt.Sprintf("Welcome to %s V%s", config.AppName, config.AppVersion)),
				)
			},
		)

		index.GET(
			"/ping", func(c *gin.Context) {
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("pong"))
			},
		)

		index.GET(
			"/alive", func(c *gin.Context) {
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("OK"))
			},
		)

		index.GET(
			"/ready", func(c *gin.Context) {
				// TODO: check dependencies (db, cache) before reporting ready
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("OK"))
			},
		)
	}

	handler := http.MaxBytesHandler(router, 8<<20)
	return &handler
}
