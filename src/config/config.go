package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName        string `required:"true" envconfig:"APP_NAME"`
	Version        string `required:"true" envconfig:"VERSION"`
	Port           string `required:"true" envconfig:"PORT"`
	MongoDBURI     string `required:"true" envconfig:"MONGO_DB_URI"`
	TestMongoDBURI string `required:"false" envconfig:"TEST_MONGO_DB_URI"`
	DBName         string `required:"true" envconfig:"DB_NAME"`
	Environment    string `required:"true" envconfig:"ENVIRONMENT"`
}

func Init() (*Config, error) {
	var cfg Config
	err := godotenv.Load("././.env")
	if err != nil {
		log.Println("Error loading .env file, falling back to cli passed env")
	}
	err = envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading environment variables: %v", err)
	}

	return &cfg, nil
}
