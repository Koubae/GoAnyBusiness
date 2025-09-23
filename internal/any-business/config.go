package any_business

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"sync"

	"github.com/Koubae/GoAnyBusiness/pkg/utils"
)

type Environment string

const (
	Testing           Environment = "testing"
	Development       Environment = "development"
	Staging           Environment = "staging"
	Production        Environment = "production"
	DefaultConfigName string      = "default"
)

var (
	Envs = [...]Environment{Testing, Development, Staging, Production}

	configLock sync.Mutex
	// Singleton mapping NOTE: Creating a map of config to make testing easier, won't hurt no one
	configsSingletonMapping = make(map[string]*Config)
)

type Config struct {
	Env            Environment
	AppName        string
	AppVersion     string
	TrustedProxies []string
	host           string
	port           uint16
}

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

	environment := Environment(utils.GetEnvString("APP_ENVIRONMENT", "development"))
	if !slices.Contains(Envs[:], environment) {
		panic(fmt.Sprintf("Invalid environment: '%s', supported envs are %v", environment, Envs))
	}
	trustedProxies := utils.GetEnvStringSlice("APP_NETWORKING_PROXIES", []string{})

	config := &Config{
		Env:            environment,
		AppName:        appName,
		AppVersion:     appVersion,
		TrustedProxies: trustedProxies,
		host:           host,
		port:           uint16(port),
	}
	configsSingletonMapping[configName] = config
	return config
}

func GetConfig(configName string) *Config {
	config, ok := configsSingletonMapping[configName]
	if !ok {
		panic(fmt.Sprintf("Config '%s' does not exist", configName))
	}
	return config
}

func GetDefaultConfig() *Config {
	return GetConfig(DefaultConfigName)
}

func (c Config) GetAddr() string {
	return fmt.Sprintf(":%d", c.port)
}

func (c Config) GetURL() string {
	return fmt.Sprintf("%s:%d", c.host, c.port)
}
