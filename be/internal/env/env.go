package env

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/shopspring/decimal"
)

func GetEnvString(key, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("Environment variable %s not set, using default value: %s", key, defaultValue)
		return defaultValue
	}
	return v
}

func GetEnvInt(key string, defaultValue int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("Environment variable %s not set, using default value: %d", key, defaultValue)
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf(
			"Error converting environment variable %s to int: %v, using default value: %d",
			key, err, defaultValue)
		return defaultValue
	}
	return n
}

func GetEnvBool(key string, defaultValue bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("Environment variable %s not set, using default value: %t", key, defaultValue)
		return defaultValue
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Printf(
			"Error converting environment variable %s to bool: %v, using default value: %t",
			key, err, defaultValue)
		return defaultValue
	}
	return b
}

func GetEnvDecimal(key string, defaultValue decimal.Decimal) decimal.Decimal {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("Environment variable %s not set, using default value: %s", key, defaultValue)
		return defaultValue
	}
	d, err := decimal.NewFromString(v)
	if err != nil {
		log.Printf(
			"Error converting environment variable %s to decimal: %v, using default value: %s",
			key, err, defaultValue)
		return defaultValue
	}
	return d
}

// mustEnv returns the value of the environment variable key, panicking when
// unset (used for required variables).
func mustEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}
