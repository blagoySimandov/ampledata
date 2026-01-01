package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	ServerAddr        string
	SerperAPIKey      string
	GroqAPIKey        string
	Crawl4aiURL       string
	GeminiAPIKey      string
	WorkersPerStage   int
	ChannelBufferSize int
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL_ENRICH", "postgres://enrichment:enrichment@localhost:5432/enrichment?sslmode=disable"),
		ServerAddr:        getEnv("SERVER_ADDR", ":8080"),
		SerperAPIKey:      getEnv("SERPER_API_KEY", ""),
		GroqAPIKey:        getEnv("GROQ_API_KEY", ""),
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		Crawl4aiURL:       getEnv("CRAWL4AI_URL", "http://localhost:8000"),
		WorkersPerStage:   getEnvInt("WORKERS_PER_STAGE", 5),
		ChannelBufferSize: getEnvInt("CHANNEL_BUFFER_SIZE", 100),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
