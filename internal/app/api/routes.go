package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/Koubae/GoAnyBusiness/internal/app/core"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ConfigureRouter configures the router
func ConfigureRouter(router *gin.Engine, config *core.Config) error {
	allowOrigin := []string{"*"}
	allowALlOrigins := false
	if config.Env != core.Production {
		allowOrigin = nil
		allowALlOrigins = true
	}

	router.Use(
		cors.New(
			cors.Config{
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
				ExposeHeaders:    []string{"Content-Length"},
				MaxAge:           12 * time.Hour,
				AllowCredentials: false,
				AllowOrigins:     allowOrigin,
				AllowAllOrigins:  allowALlOrigins,
			},
		),
	)
	err := router.SetTrustedProxies(config.TrustedProxies)
	if err != nil {
		return errors.New(fmt.Sprintf("Error setting trusted proxies, error: %s", err.Error()))
	}

	index := router.Group("/")
	indexController := &IndexController{
		config: config,
	}
	{
		index.GET("/", indexController.Index)
		index.GET("/ping", indexController.Ping)
		index.GET("/alive", indexController.Alive)
		index.GET("/ready", indexController.Ready)
	}

	return nil
}
