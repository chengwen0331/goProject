package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Access environment variables
	secretKey := os.Getenv("SECRET_KEY")
	fmt.Println("SECRET_KEY:", secretKey)
}
