package core

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"sync"

	"github.com/Koubae/GoAnyBusiness/pkg/utils"
)

// Environment represents the environment
type Environment string

// Supported environments
const (
	Testing           Environment = "testing"
	Development       Environment = "development"
	Staging           Environment = "staging"
	Production        Environment = "production"
	DefaultConfigName string      = "default"
)

var (
	// Envs is the list of supported environments
	Envs = [...]Environment{Testing, Development, Staging, Production}

	configLock sync.Mutex
	// Singleton mapping NOTE: Creating a map of config to make testing easier, won't hurt no one
	configsSingletonMapping = make(map[string]*Config)
)

// Config represents the application config
type Config struct {
	Env            Environment
	TrustedProxies []string
	host           string
	port           uint16
	AppName        string
	AppVersion     string
	AppLogLevel    string
}

// NewConfig creates a new config
func NewConfig(configName string) *Config {
	configLock.Lock()
	defer configLock.Unlock()

	_, ok := configsSingletonMapping[configName]
	if ok {
		panic(fmt.Sprintf("Config '%s' already exists", configName))
	}

	host := utils.GetEnvString("APP_HOST", "http://localhost")
	port := utils.GetEnvInt("APP_PORT", 8001)

	err := os.Setenv("PORT", strconv.Itoa(port)) // For gin-gonic
	if err != nil {
		panic(fmt.Sprintf("Error setting Gin env PORT '%v', error: %s", port, err.Error()))
	}

	appName := utils.GetEnvString("APP_NAME", "unknown")
	appVersion := utils.GetEnvString("APP_VERSION", "unknown")
	appLogLevel := utils.GetEnvString("APP_LOG_LEVEL", "INFO")

	environment := Environment(utils.GetEnvString("APP_ENVIRONMENT", "development"))
	if !slices.Contains(Envs[:], environment) {
		panic(fmt.Sprintf("Invalid environment: '%s', supported envs are %v", environment, Envs))
	}
	trustedProxies := utils.GetEnvStringSlice("APP_NETWORKING_PROXIES", []string{})

	config := &Config{
		Env:            environment,
		TrustedProxies: trustedProxies,
		host:           host,
		port:           uint16(port),
		AppName:        appName,
		AppVersion:     appVersion,
		AppLogLevel:    appLogLevel,
	}
	configsSingletonMapping[configName] = config
	return config
}

// GetConfig returns a config by name
func GetConfig(configName string) *Config {
	config, ok := configsSingletonMapping[configName]
	if !ok {
		panic(fmt.Sprintf("Config '%s' does not exist", configName))
	}
	return config
}

// GetDefaultConfig returns the default config
func GetDefaultConfig() *Config {
	return GetConfig(DefaultConfigName)
}

// GetAddr returns the address of the server
func (c Config) GetAddr() string {
	return fmt.Sprintf(":%d", c.port)
}

// GetURL returns the URL of the server
func (c Config) GetURL() string {
	return fmt.Sprintf("%s:%d", c.host, c.port)
}
