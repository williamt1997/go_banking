package initializers

import (
	"log"

	"github.com/joho/godotenv"
)

func Get_envs() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

}
