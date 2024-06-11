// Save this file in ./main.go

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"iyaem/platform/authenticator"
	"iyaem/platform/router"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	connStr := fmt.Sprintf("postgresql://postgres:%s@%s:%s/iam?sslmode=disable", os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open the database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	defer db.Close()

	rtr := router.New(auth, db)

	log.Print("Server listening on http://localhost:8080/")
	if err := rtr.Run(":8080"); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
