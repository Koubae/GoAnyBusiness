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

	"github.com/Koubae/GoAnyBusiness/internal/app/api"
	"github.com/Koubae/GoAnyBusiness/internal/app/core"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Run starts the server
func Run() {
	config := initEnv()
	logger, loggerMiddleware := core.CreateLogger(config)

	router := gin.New()
	router.Use(*loggerMiddleware, ginzap.RecoveryWithZap(logger, true)) // ref router.Use(gin.Logger(), gin.Recovery())
	api.ConfigureRouter(router, config)

	handler := http.MaxBytesHandler(router, 8<<20)
	srv := &http.Server{
		Addr:              config.GetAddr(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	srvName := fmt.Sprintf("Service %s-V%s (%s)", config.AppName, config.AppVersion, config.GetAddr())

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

func initEnv() *core.Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	config := core.NewConfig(core.DefaultConfigName)
	switch config.Env {
	case core.Testing:
		gin.SetMode(gin.TestMode)
	case core.Development, core.Staging:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	return config
}
