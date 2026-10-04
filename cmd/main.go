package main

import (
	"os"
	"webhookdelivery/internal"
	"webhookdelivery/internal/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")

	pool, err := db.Connect(databaseURL)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer pool.Close()

	app := internal.NewApplication(":8080", pool)
	if err := app.Run(); err != nil {
		panic("Application failed to run: " + err.Error())
	}
}
