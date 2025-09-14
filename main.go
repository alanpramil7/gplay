package main

import (
	"log"
	"os"

	"github.com/alanpramil7/gplay/cmd"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't treat this as a fatal error
		log.Printf("Warning: could not load .env file: %v", err)
	}

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
