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
	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/blagoySimandov/ampledata/go/internal/enricher"
	"github.com/blagoySimandov/ampledata/go/internal/pipeline"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

func main() {
	cfg := config.Load()

	store, err := state.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL store: %v", err)
	}
	defer store.Close()

	stateManager := state.NewStateManager(store)

	queryBuilder := services.NewQueryBuilder(nil, nil)
	webSearcher := services.NewSerperClient(cfg.SerperAPIKey)
	decisionMaker := services.NewGroqDecisionMaker(cfg.GroqAPIKey)
	crawler := services.NewCrawl4aiClient(cfg.Crawl4aiURL)
	extractor, err := services.NewGeminiContentExtractor(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to create Gemini content extractor: %v", err)
	}

	stages := []pipeline.Stage{
		pipeline.NewSerpStage(queryBuilder, webSearcher, stateManager, cfg.WorkersPerStage),
		pipeline.NewDecisionStage(decisionMaker, stateManager, cfg.WorkersPerStage),
		pipeline.NewCrawlStage(crawler, stateManager, cfg.WorkersPerStage),
		pipeline.NewExtractStage(extractor, stateManager, cfg.WorkersPerStage),
	}

	pipelineConfig := &pipeline.PipelineConfig{
		WorkersPerStage:   cfg.WorkersPerStage,
		ChannelBufferSize: cfg.ChannelBufferSize,
	}
	p := pipeline.NewPipeline(stateManager, stages, pipelineConfig)

	enr := enricher.NewEnricher(p, stateManager)

	// Initialize WorkOS client for authentication
	workosClient := usermanagement.NewClient(cfg.WorkOSAPIKey)

	handler := api.NewEnrichHandler(enr)
	router := api.SetupRoutes(handler, workosClient)

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
