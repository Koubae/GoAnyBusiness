package core

import (
	"fmt"
	"strings"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const DefaultLoggerName = "default"

var loggerSingleton = make(map[string]*zap.SugaredLogger)

// CreateLogger creates a new logger
func CreateLogger(config *Config) (*zap.Logger, *gin.HandlerFunc, error) {
	var cnf *zap.Config
	level := parseLogLevel(config.AppLogLevel)

	switch config.Env {
	case Testing, Development:
		cnf = newDevelopmentConfig(level)
	default:
		cnf = newProductionConfig(level)
	}

	logger, err := cnf.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return nil, nil, fmt.Errorf("build logger error: %w", err)
	}

	middleware := ginzap.GinzapWithConfig(
		logger,
		&ginzap.Config{TimeFormat: time.RFC3339, UTC: true, DefaultLevel: zapcore.InfoLevel},
	)
	loggerSingleton[DefaultLoggerName] = logger.Sugar()
	return logger, &middleware, nil
}

// GetLogger returns a logger by name
func GetLogger(name string) *zap.SugaredLogger {
	logger, ok := loggerSingleton[name]
	if !ok {
		panic(fmt.Sprintf("Logger '%s' does not exist", name))
	}
	return logger
}

// GetDefaultLogger returns the default logger
func GetDefaultLogger() *zap.SugaredLogger {
	return GetLogger(DefaultLoggerName)
}

func newProductionConfig(level zapcore.Level) *zap.Config {
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

func newDevelopmentConfig(level zapcore.Level) *zap.Config {
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
