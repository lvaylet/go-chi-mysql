package main

func main() {
	a := App{}

	// Initialize app with username, password and database name
	a.Initialize("DB_USERNAME", "DB_PASSWORD", "rest_api_example")

	a.Run(":8080")
}
