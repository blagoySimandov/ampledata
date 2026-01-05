package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	FE_BASE_URL       string
	ServerAddr        string
	SerperAPIKey      string
	GroqAPIKey        string
	Crawl4aiURL       string
	GeminiAPIKey      string
	WorkersPerStage   int
	ChannelBufferSize int
	WorkOSApiKey      string
	WorkOSClientID    string
	WorkOSRedirectURL string
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL_ENRICH", "postgres://enrichment:enrichment@localhost:5432/enrichment?sslmode=disable"),
		FE_BASE_URL:       getEnv("FE_BASE_URL", "http://localhost:5173"),
		ServerAddr:        getEnv("SERVER_ADDR", ":8080"),
		SerperAPIKey:      getEnv("SERPER_API_KEY", ""),
		GroqAPIKey:        getEnv("GROQ_API_KEY", ""),
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		Crawl4aiURL:       getEnv("CRAWL4AI_URL", "http://localhost:8000"),
		WorkersPerStage:   getEnvInt("WORKERS_PER_STAGE", 5),
		ChannelBufferSize: getEnvInt("CHANNEL_BUFFER_SIZE", 100),
		WorkOSApiKey:      getEnv("WORKOS_API_KEY", ""),
		WorkOSClientID:    getEnv("WORKOS_CLIENT_ID", ""),
		WorkOSRedirectURL: getEnv("WORKOS_REDIRECT_URL", "http://localhost:8080/auth/callback"),
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

var config *Config

func GetConfig() *Config {
	if config == nil {
		config = Load()
	}
	return config
}
