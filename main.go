package main

import (
	"fmt"
	"os"
)

func main() {
	a := App{}
	a.Initialize(
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	port := getEnvWithDefault("LISTEN_ON_PORT", "8080")
	a.Run(fmt.Sprintf(":%s", port))
}

func getEnvWithDefault(key, defaultValue string) string {
	value, isDefined := os.LookupEnv(key)
	if isDefined {
		return value
	}
	return defaultValue
}
