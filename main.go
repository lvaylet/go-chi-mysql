package main

func main() {
	a := App{}

	// Initialize app with username, password and database name
	a.Initialize("root", "UN5unZT2YMxRaPR", "rest_api_example")

	a.Run(":8080")
}
`