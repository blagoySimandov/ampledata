package api

import (
	"github.com/blagoySimandov/ampledata/go/internal/enricher"
	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/blagoySimandov/ampledata/go/internal/user"
)

type Server struct {
	enricher    enricher.IEnricher
	gcsReader   *gcs.CSVReader
	store       state.Store
	userRepo    user.Repository
	billing     services.BillingService
	keySelector services.KeySelector
}

func NewServer(
	enr enricher.IEnricher,
	gcsReader *gcs.CSVReader,
	store state.Store,
	userRepo user.Repository,
	billing services.BillingService,
	keySelector services.KeySelector,
) *Server {
	return &Server{
		enricher:    enr,
		gcsReader:   gcsReader,
		store:       store,
		userRepo:    userRepo,
		billing:     billing,
		keySelector: keySelector,
	}
}
