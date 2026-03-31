package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	ServerAddr        string
	SerperAPIKey      string
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
	MaxEnrichmentRetries          int
	MaxOrganicResults             int
	ConcurrencyRowEnrichmentLimit int

	CreditsPerCell int

	SerperCost              int
	TknInCost               int
	TknOutCost              int
	StripeSecretKey         string
	StripeWebhookSecret     string
	EnrichmentCostMeterName string

	StaticDir string

	// Stripe Price IDs – populated by Terraform, set as environment variables.
	// Run `terraform apply` in terraform/stripe and export the outputs.
	StarterBasePriceID       string
	StarterMeteredPriceID    string
	ProBasePriceID           string
	ProMeteredPriceID        string
	EnterpriseBasePriceID    string
	EnterpriseMeteredPriceID string

	FreeTierCredits int64
}

const (
	StripeMetadataTier        = "ampledata_tier"
	StripeMetadataProductType = "ampledata_product_type"
	StripeMetadataPriceType   = "ampledata_price_type"
	StripePriceTypeBase       = "base"
	StripePriceTypeMetered    = "metered"
	StripeTierIDKey           = "tier_id"
)

var cfg Config = Config{
	DatabaseURL:       getEnv("DATABASE_URL_ENRICH", "postgres://enrichment:enrichment@localhost:5432/enrichment?sslmode=disable"),
	ServerAddr:        getEnv("SERVER_ADDR", ":8080"),
	SerperAPIKey:      getEnv("SERPER_API_KEY", ""),
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
	MaxEnrichmentRetries:          getEnvInt("MAX_ENRICHMENT_RETRIES", 2),
	MaxOrganicResults:             getEnvInt("MAX_ORGANIC_RESULTS", 4),
	ConcurrencyRowEnrichmentLimit: getEnvInt("CONCURRENCY_ROW_ENRICHMENT_LIMIT", 10),

	CreditsPerCell: getEnvInt("CREDITS_PER_CELL", 1),

	// Token costs are stored in nano-dollars per token (billionths of a dollar).
	// To convert: price_per_million_tokens * 1000 = nano_dollars_per_token
	// e.g. $0.25/1M input → 250 nano-dollars/token
	//
	// Reference pricing (per 1M tokens):
	//   gemini-2.5-flash-lite-preview-09-2025:  in $0.10  → 100,  out $0.40  → 400
	//   gemini-3.1-flash-lite-preview:          in $0.25  → 250,  out $1.50  → 1500
	//   gemini-2.5-flash:                       in 0.30 -> 300,  out $2.50  → 2500
	SerperCost:              getEnvInt("SERPER_COST", 1),
	TknInCost:               getEnvInt("TKN_INGESTION_COST", 300),
	TknOutCost:              getEnvInt("TKN_ENRICHMENT_COST", 2500),
	StripeSecretKey:         getEnv("STRIPE_SECRET", ""),
	StripeWebhookSecret:     getEnv("STRIPE_WEBHOOK_SECRET", ""),
	EnrichmentCostMeterName: getEnv("ENRICHMENT_COST_METER_NAME", "enrichment_credits"),

	StaticDir: getEnv("STATIC_DIR", "web/ampledata-fe/dist"),

	StarterBasePriceID:       getEnv("STRIPE_STARTER_BASE_PRICE_ID", ""),
	StarterMeteredPriceID:    getEnv("STRIPE_STARTER_METERED_PRICE_ID", ""),
	ProBasePriceID:           getEnv("STRIPE_PRO_BASE_PRICE_ID", ""),
	ProMeteredPriceID:        getEnv("STRIPE_PRO_METERED_PRICE_ID", ""),
	EnterpriseBasePriceID:    getEnv("STRIPE_ENTERPRISE_BASE_PRICE_ID", ""),
	EnterpriseMeteredPriceID: getEnv("STRIPE_ENTERPRISE_METERED_PRICE_ID", ""),

	FreeTierCredits: int64(getEnvInt("FREE_TIER_CREDITS", 100)),
}

func Load() *Config {
	return &cfg
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
