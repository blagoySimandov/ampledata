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
	WorkOSAPIKey      string
	WorkOSClientID    string
	WorkersPerStage   int
	ChannelBufferSize int
	DebugAuthBypass   bool

	// Temporal configuration
	TemporalHostPort  string
	TemporalNamespace string
	TemporalTaskQueue string

	// Enrichment configuration
	MaxEnrichmentRetries int
	// CostTracking
	SerperCost         int
	TknInCost          int
	TknOutCost         int
	CreditExchangeRate int
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL_ENRICH", "postgres://enrichment:enrichment@localhost:5432/enrichment?sslmode=disable"),
		ServerAddr:        getEnv("SERVER_ADDR", ":8080"),
		SerperAPIKey:      getEnv("SERPER_API_KEY", ""),
		GroqAPIKey:        getEnv("GROQ_API_KEY", ""),
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		WorkOSAPIKey:      getEnv("WORKOS_API_KEY", ""),
		WorkOSClientID:    getEnv("WORKOS_CLIENT_ID", ""),
		Crawl4aiURL:       getEnv("CRAWL4AI_URL", "http://localhost:8000"),
		WorkersPerStage:   getEnvInt("WORKERS_PER_STAGE", 5),
		ChannelBufferSize: getEnvInt("CHANNEL_BUFFER_SIZE", 100),
		DebugAuthBypass:   getEnvBool("DEBUG_AUTH_BYPASS", false),

		// Temporal settings
		TemporalHostPort:  getEnv("TEMPORAL_HOST_PORT", "localhost:7233"),
		TemporalNamespace: getEnv("TEMPORAL_NAMESPACE", "default"),
		TemporalTaskQueue: getEnv("TEMPORAL_TASK_QUEUE", "ampledata-enrichment"),

		// Enrichment settings
		MaxEnrichmentRetries: getEnvInt("MAX_ENRICHMENT_RETRIES", 1),

		// CostTracking
		SerperCost:         getEnvInt("SERPER_COST", 0),
		TknInCost:          getEnvInt("TKN_INGESTION_COST", 500),
		TknOutCost:         getEnvInt("TKN_ENRICHMENT_COST", 3000),
		CreditExchangeRate: getEnvInt("CREDIT_EXCHANGE_RATE", 0),
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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
