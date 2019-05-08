package main

import (
	"fmt"
	"os"
)

func main() {
	a := App{}
	a.Initialize(
		os.Getenv("CLOUD_SQL_DB_USERNAME"),
		os.Getenv("CLOUD_SQL_DB_PASSWORD"),
		os.Getenv("CLOUD_SQL_INSTANCE_CONNECTION_NAME"),
		os.Getenv("CLOUD_SQL_DB_NAME"))

	port := getEnvWithDefault("PORT", "8080")
	a.Run(fmt.Sprintf(":%s", port))
}

func getEnvWithDefault(key, defaultValue string) string {
	value, isDefined := os.LookupEnv(key)
	if isDefined {
		return value
	}
	return defaultValue
}
