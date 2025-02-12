package config

import (
	"log"
	"os"
)

type Config struct {
	BotToken       string
	UniqueServerID string
}

var GlobalConfig *Config

func LoadConfig() *Config {
	GlobalConfig = &Config{
		BotToken:       getEnv("BOT_TOKEN", ""),
		UniqueServerID: getEnv("UNIQUE_SERVER_ID", ""),
	}
	return GlobalConfig
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	if fallback == "" {
		log.Fatalf("Environment variable %s not set and no fallback provided", key)
	}
	return fallback
}
