// Save this file in ./main.go

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
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

	pubSub, err := providers.NewPubSub(os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		log.Fatalf("Failed to initialize the pubsub client: %v", err)
	}

	defer pubSub.CloseConnection()

	callbacks := make([]providers.Callback, 0)
	callbacks = append(callbacks, printMsg)

	go pubSub.Subscribe("iam_domain_registered", callbacks)

	log.Print("Server listening on http://localhost:8080/")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}

}

func printMsg(ctx context.Context, msg *pubsub.Message) {

	var messageJson map[string]interface{}

	json.Unmarshal(msg.Data, &messageJson)

	log.Printf("Message: %v", messageJson)
}
