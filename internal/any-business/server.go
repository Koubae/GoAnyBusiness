package any_business

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
		panic(err.Error())
	}

	index := router.Group("/")
	{
		index.GET(
			"/", func(c *gin.Context) {
				c.Data(
					200,
					"text/html; charset=utf-8",
					[]byte(fmt.Sprintf("Welcome to %s V%s", config.AppName, config.AppVersion)),
				)
			},
		)

		index.GET(
			"/ping", func(c *gin.Context) {
				c.Data(200, "text/html; charset=utf-8", []byte("pong"))
			},
		)

		index.GET(
			"/alive", func(c *gin.Context) {
				c.Data(http.StatusNoContent, "text/html; charset=utf-8", []byte("OK"))
			},
		)

		index.GET(
			"/ready", func(c *gin.Context) {
				// TODO: check dependencies (db, cache) before reporting ready
				c.Data(http.StatusNoContent, "text/html; charset=utf-8", []byte("OK"))
			},
		)
	}

	srvName := fmt.Sprintf("Service %s-V%s (%s)", config.AppName, config.AppVersion, config.GetAddr())
	srv := &http.Server{
		Addr:              config.GetAddr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	shutdownErr := make(chan error, 1)
	go func() {
		log.Printf("%s | Server starting...\n", srvName)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			shutdownErr <- fmt.Errorf("%s - Error while shutting down server, error: %v\n", srvName, err)
			return
		}
		shutdownErr <- nil

	}()
	log.Printf("%s | Server started\n", srvName)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	select {
	case <-ctx.Done():
		sig := <-sigCh
		log.Printf(
			"%s - Server Shutting down gracefully (received signal: '%s'), press Ctrl+C again to force\n",
			srvName,
			sig,
		)
	case err := <-shutdownErr:
		if err != nil {
			log.Fatalf("%s - server startup/runtime failure, error: %v\n", srvName, err) // startup/runtime failure
		}

		log.Printf(
			"%s - Server Shutting down gracefully (After server stop serving), press Ctrl+C again to force\n",
			srvName,
		)
	}

	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close() // If shutdown times out, force close:
		log.Fatalf("%s - Server forced to shutdown: %v\n", srvName, err)
	}

	log.Printf("%s - Server Shutdown, cleaning up resources\n", srvName)
	// TODO: cleanup resources
	log.Printf("%s - Server exiting\n", srvName)

}

func initEnv() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %s", err.Error()))
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
