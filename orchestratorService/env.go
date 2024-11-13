package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func initializeEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file, exiting....")
	}
}

func checkIfAllEnvVarsAreSet() {
	if os.Getenv("DOCKER_RUNNER_SERVICE") == "" {
		log.Fatal("DOCKER_RUNNER_SERVICE environment variable is not set, exiting....")
	} else if os.Getenv("CONFIG_FILE_PATH") == "" {
		log.Fatal("CONFIG_FILE_PATH environment variable is not set, exiting....")
	}
}
