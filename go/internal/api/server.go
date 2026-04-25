package api

import (
	"github.com/blagoySimandov/ampledata/go/internal/enricher"
	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/googleoauth"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/sheets"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/blagoySimandov/ampledata/go/internal/user"
)

type Server struct {
	enricher       enricher.IEnricher
	gcsReader      *gcs.CSVReader
	store          state.Store
	userRepo       user.Repository
	billing        services.BillingService
	keySelector    services.KeySelector
	sourcesService services.ISourcesService
	oauthService   *googleoauth.Service
	sheetsClient   *sheets.Client
}

func NewServer(
	enr enricher.IEnricher,
	gcsReader *gcs.CSVReader,
	store state.Store,
	userRepo user.Repository,
	billing services.BillingService,
	keySelector services.KeySelector,
	aiclient services.IAIClient,
	sourcesService services.ISourcesService,
	oauthService *googleoauth.Service,
	sheetsClient *sheets.Client,
) *Server {
	return &Server{
		enricher:       enr,
		gcsReader:      gcsReader,
		store:          store,
		userRepo:       userRepo,
		billing:        billing,
		keySelector:    keySelector,
		sourcesService: sourcesService,
		oauthService:   oauthService,
		sheetsClient:   sheetsClient,
	}
}
