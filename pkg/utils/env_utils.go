package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetEnvString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		val = strings.TrimSpace(val)
		if val != "" {
			return val
		}
	}
	return defaultVal

}

func GetEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if valInt, err := strconv.Atoi(val); err != nil {
			panic(fmt.Sprintf("Invalid value int for %s: %s", key, val))
		} else {
			return valInt
		}
	}
	return defaultVal
}

func GetEnvBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if valBool, err := strconv.ParseBool(val); err != nil {
			panic(fmt.Sprintf("Invalid value bool for %s: %s", key, val))
		} else {
			return valBool
		}
	}

	return defaultVal
}

func GetEnvStringSlice(key string, defaultVal []string) []string {
	if val, ok := os.LookupEnv(key); ok {
		items := strings.Split(val, ",")
		itemsTrimmed := make([]string, 0, len(items))
		for _, item := range items {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				itemsTrimmed = append(itemsTrimmed, trimmed)
			}
		}
		if len(itemsTrimmed) > 0 {
			return itemsTrimmed
		}
	}
	return defaultVal
}

func GetEnvIntSlice(key string, defaultVal []int) []int {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		items := strings.Split(val, ",")
		itemsInt := make([]int, 0, len(items))
		for _, item := range items {
			if val, err := strconv.Atoi(strings.TrimSpace(item)); err != nil {
				panic(err.Error())
			} else {
				itemsInt = append(itemsInt, val)
			}
		}

		if len(itemsInt) > 0 {
			return itemsInt
		}
	}

	return defaultVal
}
