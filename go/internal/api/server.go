package api

import "github.com/blagoySimandov/ampledata/go/internal/gcs"

type Server struct {
	enricher       IEnricher
	gcsReader      *gcs.CSVReader
	store          Store
	userRepo       UserRepo
	templatesRepo  ITemplateRepo
	billing        BillingService
	keySelector    KeySelector
	sourcesService SourcesService
}

func NewServer(
	enr IEnricher,
	gcsReader *gcs.CSVReader,
	store Store,
	userRepo UserRepo,
	billing BillingService,
	keySelector KeySelector,
	sourcesService SourcesService,
	templatesRepo ITemplateRepo,
) *Server {
	return &Server{
		enricher:       enr,
		gcsReader:      gcsReader,
		store:          store,
		userRepo:       userRepo,
		billing:        billing,
		keySelector:    keySelector,
		sourcesService: sourcesService,
		templatesRepo:  templatesRepo,
	}
}
