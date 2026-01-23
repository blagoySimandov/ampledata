package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/api"
	"github.com/blagoySimandov/ampledata/go/internal/auth"
	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/blagoySimandov/ampledata/go/internal/db"
	"github.com/blagoySimandov/ampledata/go/internal/enricher"
	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/activities"
	temporalClient "github.com/blagoySimandov/ampledata/go/internal/temporal/client"
	"github.com/blagoySimandov/ampledata/go/internal/temporal/worker"
	"github.com/blagoySimandov/ampledata/go/internal/user"
)

func main() {
	cfg := config.Load()

	//_ := stripe.NewClient(cfg.StripeSecretKey)

	db := db.NewBunPostgresClient(cfg.DatabaseURL)
	store, err := state.NewPostgresStore(db)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL store: %v", err)
	}
	defer store.Close()

	userRepo := user.NewUserRepository(db)
	if err := userRepo.InitializeDatabase(context.Background()); err != nil {
		log.Fatalf("Failed to initialize user database: %v", err)
	}

	costTracker, err := services.NewCostTracker(cfg.TknInCost, cfg.TknOutCost, cfg.SerperCost, cfg.CreditExchangeRate, services.WithStore(store))
	if err != nil {
		log.Fatalf("Failed to create cost tracker: %v", err)
	}

	gcsReader, err := gcs.NewCSVReader("ampledata-enrichment-uploads")
	if err != nil {
		log.Fatalf("Failed to create GCS reader: %v", err)
	}
	defer gcsReader.Close()
	stateManager := state.NewStateManager(store)

	aiClient, err := services.NewGeminiAIClient(services.WithCostTracker(costTracker))
	if err != nil {
		log.Fatalf("Failed to create AI client: %v", err)
	}
	patternGenerator, err := services.NewPatternGenerator(aiClient)
	if err != nil {
		log.Fatalf("Failed to create Gemini pattern generator: %v", err)
	}
	webSearcher := services.NewSerperClient(cfg.SerperAPIKey)
	decisionMaker, err := services.NewGeminiDecisionMaker(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to create Gemini decision maker: %v", err)
	}
	crawler := services.NewCrawl4aiClient(cfg.Crawl4aiURL)
	extractor, err := services.NewAIContentExtractor(aiClient)
	if err != nil {
		log.Fatalf("Failed to create Gemini content extractor: %v", err)
	}
	keySelector, err := services.NewGeminiKeySelector(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to create Gemini key selector: %v", err)
	}
	tc, err := temporalClient.NewClient(cfg.TemporalHostPort, cfg.TemporalNamespace)
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer tc.Close()

	acts := activities.NewActivities(
		stateManager,
		webSearcher,
		decisionMaker,
		crawler,
		extractor,
		patternGenerator,
	)

	w := worker.NewWorker(tc, cfg.TemporalTaskQueue, acts)
	err = w.Start()
	if err != nil {
		log.Fatalf("Failed to start Temporal worker: %v", err)
	}
	defer w.Stop()

	// Create Temporal-based enricher
	enr := enricher.NewTemporalEnricher(tc, stateManager, cfg.TemporalTaskQueue, cfg.MaxEnrichmentRetries)

	jwtVerifier, err := auth.NewJWTVerifier(cfg.WorkOSClientID, cfg.DebugAuthBypass)
	if err != nil {
		log.Fatalf("Failed to create JWT verifier: %v", err)
	}

	handler := api.NewEnrichHandler(enr, gcsReader, store)
	keySelectorHandler := api.NewKeySelectorHandler(keySelector, gcsReader, store)
	router := api.SetupRoutes(handler, keySelectorHandler, jwtVerifier)

	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on %s", cfg.ServerAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server stopped")
}
