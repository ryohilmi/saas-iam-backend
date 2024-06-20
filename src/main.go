// Save this file in ./main.go

package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"iyaem/internal/presentation/routes"
	"iyaem/internal/providers"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := providers.NewAuthenticator()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	dbConfig := providers.DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	}

	db, err := providers.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize the database: %v", err)
	}

	defer db.Close()

	router := routes.NewRouter(auth, db)

	log.Print("Server listening on http://localhost:8080/")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
