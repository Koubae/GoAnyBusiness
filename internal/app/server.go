package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Run starts the server
func Run() {
	config := initEnv()

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
				AllowAllOrigins:  config.Env != Production,
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

	srvName := fmt.Sprintf("Service %s-V%s (%s)", config.AppName, config.AppVersion, config.GetAddr())
	handler := http.MaxBytesHandler(router, 8<<20) // server: cap request body size (e.g., 8 MiB)
	srv := &http.Server{
		Addr:              config.GetAddr(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	startUpErr := make(chan error, 1)
	go func() {
		log.Printf("%s | Server starting...", srvName)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			startUpErr <- fmt.Errorf("server issues while listening: %v", err)
			return
		}
		startUpErr <- nil
	}()
	log.Printf("%s | Server started", srvName)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(sigCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := <-sigCh
		log.Printf("%s - shutting down gracefully (received signal: %s); press Ctrl+C again to force", srvName, sig)
		cancel()
	}()

	select {
	case <-ctx.Done():
	case err := <-startUpErr:
		if err != nil {
			log.Printf("%s - server startup/runtime failure, error: %v", srvName, err) // startup/runtime failure
			return
		}
		log.Printf(
			"%s - Server Shutting down gracefully (After server stop serving), press Ctrl+C again to force",
			srvName,
		)

	}

	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close() // If shutdown times out, force close:
		log.Printf("%s - Server forced to shutdown: %v", srvName, err)
		return
	}

	log.Printf("%s - Server Shutdown, cleaning up resources", srvName)
	// TODO: cleanup resources
	log.Printf("%s - Server exiting", srvName)
}

func initEnv() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	config := NewConfig(DefaultConfigName)
	switch config.Env {
	case Testing:
		gin.SetMode(gin.TestMode)
	case Development, Staging:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	return config
}
