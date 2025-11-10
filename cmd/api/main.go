package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mlhmz/dockermc-cloud-manager/internal/cli/commands"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
