package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Koubae/GoAnyBusiness/internal/app/api"
	"github.com/Koubae/GoAnyBusiness/internal/app/core"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Run starts the server
func Run() {
	config := initEnv()
	logger, loggerMiddleware := createLogger(config)

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

func NewProductionConfig(level zapcore.Level) *zap.Config {
	return &zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func NewDevelopmentConfig(level zapcore.Level) *zap.Config {
	return &zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			// Keys can be anything except the empty string.
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			FunctionKey:    zapcore.OmitKey,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func parseLogLevel(s string) zapcore.Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN", "WARNING":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "DPANIC":
		return zapcore.DPanicLevel
	case "PANIC":
		return zapcore.PanicLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func createLogger(config *core.Config) (*zap.Logger, *gin.HandlerFunc) {
	var cnf *zap.Config
	level := parseLogLevel(config.AppLogLevel)

	switch config.Env {
	case core.Testing, core.Development:
		cnf = NewDevelopmentConfig(level)
	default:
		cnf = NewProductionConfig(level)
	}

	logger, _ := cnf.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	middleware := ginzap.GinzapWithConfig(
		logger,
		&ginzap.Config{TimeFormat: time.RFC3339, UTC: true, DefaultLevel: zapcore.InfoLevel},
	)
	return logger, &middleware
}
