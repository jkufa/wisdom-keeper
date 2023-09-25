package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AuthToken     string
	Environment   string
	ModChannelId  string
	LogChannelId  string
	CoolDownHours int
}

func NewConfig() *Config {
	return &Config{
		AuthToken:     "",
		Environment:   "",
		ModChannelId:  "",
		LogChannelId:  "",
		CoolDownHours: 0,
	}
}

func GetConfig() *Config {
	LoadEnv()

	cd := os.Getenv("COOLDOWN_HOURS")
	cdInt, err := strconv.Atoi(cd)
	if err != nil {
		log.Fatal(err)
	}

	return &Config{
		AuthToken:     os.Getenv("AUTH_TOKEN"),
		Environment:   os.Getenv("ENV"),
		ModChannelId:  os.Getenv("MOD_CHANNEL_ID"),
		LogChannelId:  os.Getenv("LOG_CHANNEL_ID"),
		CoolDownHours: cdInt,
	}
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
