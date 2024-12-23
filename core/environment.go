package core

import (
	"os"
	"strconv"
)

func GetStringVariable(key string, defaultValue string) string {
	result, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return result
}

func GetIntVariable(key string, defaultValue int) int {
	tmp, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	result, err := strconv.Atoi(tmp)
	if err != nil {
		return defaultValue
	}

	return result
}

func GetBoolVariable(key string, defaultValue bool) bool {
	tmp, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	result := tmp == "1"
	return result
}
