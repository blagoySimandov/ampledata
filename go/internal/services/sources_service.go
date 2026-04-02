package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/google/uuid"
)

type IEnricher interface {
	Enrich(ctx context.Context, jobID, userID, stripeCustomerID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription *string) error
}

// ICSVReader abstracts file reading operations so the service can be unit-tested
// without a real GCS connection.
type ICSVReader interface {
	ReadCSV(ctx context.Context, objectName string) (*gcs.CSVResult, error)
	ReadCompositeKeyFromFile(ctx context.Context, objectName string, columnNames []string) ([]string, error)
	ReadCompositeKeyFromFileFiltered(ctx context.Context, objectName string, keyColumns []string, filterColumns []string) ([]string, error)
	GenerateSignedURL(ctx context.Context, objectName, contentType string) (string, error)
}

var (
	ErrSourceNotFound      = errors.New("source not found")
	ErrSourceForbidden     = errors.New("forbidden")
	ErrInsufficientCredits = errors.New("insufficient credits to run this job")
)

type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

func newValidationError(msg string) error { return ValidationError{Msg: msg} }

type SourceWithJobs struct {
	Source *models.Source
	Jobs   []*models.Job
}

type EnrichSourceInput struct {
	SourceID             uuid.UUID
	AuthUserID           string
	DBUser               *models.User
	KeyColumns           []string
	KeyColumnDescription *string
	ColumnsMetadata      []*models.ColumnMetadata
	MaxRows              *int
}

type ISourcesService interface {
	ListSources(ctx context.Context, userID string, offset, limit int) ([]*SourceWithJobs, error)
	GetSource(ctx context.Context, sourceID uuid.UUID, userID string) (*SourceWithJobs, error)
	GetSourceData(ctx context.Context, sourceID uuid.UUID, userID string) (*gcs.CSVResult, error)
	EnrichSource(ctx context.Context, input EnrichSourceInput) (string, error)
	CreateUploadSource(ctx context.Context, userID, contentType string, headers []string) (uuid.UUID, string, error)
}

type ISourceNameGeneratorPromptService interface {
	GenerateSourceNamePrompt(ctx context.Context, headers []string) string
}

type sourcesService struct {
	store         state.Store
	reader        ICSVReader
	enricher      IEnricher
	aiClient      IAIClient
	promptService ISourceNameGeneratorPromptService
}

func NewSourcesService(store state.Store, reader ICSVReader, enr IEnricher, aiclient IAIClient, promptService ISourceNameGeneratorPromptService) ISourcesService {
	return &sourcesService{store: store, reader: reader, enricher: enr, aiClient: aiclient, promptService: promptService}
}

func (s *sourcesService) ListSources(ctx context.Context, userID string, offset, limit int) ([]*SourceWithJobs, error) {
	sources, err := s.store.GetSourcesByUser(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}
	return s.attachJobs(ctx, sources)
}

func (s *sourcesService) attachJobs(ctx context.Context, sources []*models.Source) ([]*SourceWithJobs, error) {
	result := make([]*SourceWithJobs, len(sources))
	for i, src := range sources {
		jobs, err := s.store.GetJobsBySource(ctx, src.ID)
		if err != nil {
			return nil, err
		}
		result[i] = &SourceWithJobs{Source: src, Jobs: jobs}
	}
	return result, nil
}

func (s *sourcesService) GetSource(ctx context.Context, sourceID uuid.UUID, userID string) (*SourceWithJobs, error) {
	source, err := s.store.GetSource(ctx, sourceID)
	if err != nil {
		return nil, ErrSourceNotFound
	}
	if source.UserID != userID {
		return nil, ErrSourceForbidden
	}
	jobs, err := s.store.GetJobsBySource(ctx, source.ID)
	if err != nil {
		return nil, err
	}
	return &SourceWithJobs{Source: source, Jobs: jobs}, nil
}

func (s *sourcesService) GetSourceData(ctx context.Context, sourceID uuid.UUID, userID string) (*gcs.CSVResult, error) {
	source, err := s.store.GetSource(ctx, sourceID)
	if err != nil {
		return nil, ErrSourceNotFound
	}
	if source.UserID != userID {
		return nil, ErrSourceForbidden
	}
	csvMeta, ok := source.Metadata.(*models.CSVSourceMetadata)
	if !ok {
		return nil, fmt.Errorf("source metadata not found")
	}
	return s.reader.ReadCSV(ctx, csvMeta.FileURI)
}

func (s *sourcesService) CreateUploadSource(ctx context.Context, userID, contentType string, headers []string) (uuid.UUID, string, error) {
	ext, _ := mime.ExtensionsByType(contentType)
	extension := ".csv"
	if len(ext) > 0 {
		extension = ext[0]
	}
	fileID := generateJobID(extension)
	source, err := s.createCSVSource(ctx, userID, fileID, contentType, headers)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to create source: %w", err)
	}
	url, err := s.reader.GenerateSignedURL(ctx, fileID, contentType)
	if err != nil {
		return uuid.Nil, "", err
	}
	return source.ID, url, nil
}

func (s *sourcesService) generateSourceName(ctx context.Context, headers []string) string {
	prompt := s.promptService.GenerateSourceNamePrompt(ctx, headers)
	name, err := s.aiClient.GenerateContent(ctx, prompt)
	if err != nil {
		return "Uploaded File"
	}
	return name
}

func (s *sourcesService) createCSVSource(ctx context.Context, userID, fileURI, contentType string, headers []string) (*models.SourceDB, error) {
	name := s.generateSourceName(ctx, headers)
	metaJSON, err := json.Marshal(&models.CSVSourceMetadata{FileURI: fileURI, ContentType: contentType, Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	now := time.Now()
	source := &models.SourceDB{
		UserID:    userID,
		Type:      models.SourceTypeCSVUpload,
		Metadata:  json.RawMessage(metaJSON),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateSource(ctx, source); err != nil {
		return nil, err
	}
	return source, nil
}

func (s *sourcesService) EnrichSource(ctx context.Context, input EnrichSourceInput) (string, error) {
	event := logger.FromContext(ctx)

	source, err := s.getOwnedSource(ctx, input.SourceID, input.AuthUserID)
	if err != nil {
		return "", err
	}

	if event != nil {
		event.StartStageByName("resolve_key_columns")
	}
	keyColumns, keyColumnDesc, err := s.resolveKeyColumns(ctx, input.SourceID, input.KeyColumns, input.KeyColumnDescription)
	if err != nil {
		if event != nil {
			event.FailStageByName("resolve_key_columns", err)
		}
		return "", err
	}
	if event != nil {
		event.CompleteStageByName("resolve_key_columns", map[string]interface{}{
			"key_columns_count": len(keyColumns),
		})
	}

	csvMeta, ok := source.Metadata.(*models.CSVSourceMetadata)
	if !ok {
		return "", fmt.Errorf("source metadata not found")
	}

	if event != nil {
		event.StartStageByName("read_row_keys")
	}
	rowKeys, err := s.readRowKeys(ctx, csvMeta.FileURI, keyColumns, input.ColumnsMetadata)
	if err != nil {
		if event != nil {
			event.FailStageByName("read_row_keys", err)
		}
		return "", newValidationError(fmt.Sprintf("failed to read CSV: %v", err))
	}
	if event != nil {
		event.CompleteStageByName("read_row_keys", map[string]interface{}{
			"row_keys_count": len(rowKeys),
		})
	}

	if len(rowKeys) == 0 {
		return "", newValidationError("no rows found in key column")
	}
	if input.MaxRows != nil && *input.MaxRows > 0 && len(rowKeys) > *input.MaxRows {
		rowKeys = rowKeys[:*input.MaxRows]
	}
	if !input.DBUser.CanEnrichCells(int64(len(rowKeys) * len(input.ColumnsMetadata))) {
		return "", ErrInsufficientCredits
	}

	if event != nil {
		event.SetMetadata("effective_rows", len(rowKeys))
		event.SetMetadata("total_cells", len(rowKeys)*len(input.ColumnsMetadata))
	}

	return s.createAndStartJob(ctx, input, keyColumns, keyColumnDesc, rowKeys)
}

func (s *sourcesService) getOwnedSource(ctx context.Context, sourceID uuid.UUID, userID string) (*models.Source, error) {
	source, err := s.store.GetSource(ctx, sourceID)
	if err != nil {
		return nil, ErrSourceNotFound
	}
	if source.UserID != userID {
		return nil, ErrSourceForbidden
	}
	return source, nil
}

func (s *sourcesService) resolveKeyColumns(ctx context.Context, sourceID uuid.UUID, keyColumns []string, desc *string) ([]string, *string, error) {
	if len(keyColumns) > 0 {
		return keyColumns, desc, nil
	}
	jobs, err := s.store.GetJobsBySource(ctx, sourceID)
	if err != nil || len(jobs) == 0 {
		return nil, nil, newValidationError("key_columns required for first enrichment run")
	}
	most := jobs[0]
	return most.KeyColumns, most.KeyColumnDescription, nil
}

func (s *sourcesService) readRowKeys(ctx context.Context, fileURI string, keyColumns []string, cols []*models.ColumnMetadata) ([]string, error) {
	imputationCols := imputationColumnNames(cols)
	var keys []string
	var err error
	if len(imputationCols) > 0 {
		keys, err = s.reader.ReadCompositeKeyFromFileFiltered(ctx, fileURI, keyColumns, imputationCols)
	} else {
		keys, err = s.reader.ReadCompositeKeyFromFile(ctx, fileURI, keyColumns)
	}
	if err != nil {
		return nil, err
	}
	return deduplicateKeys(keys), nil
}

func (s *sourcesService) createAndStartJob(ctx context.Context, input EnrichSourceInput, keyColumns []string, keyColumnDesc *string, rowKeys []string) (string, error) {
	jobID := generateJobID(".csv")
	if err := s.store.CreatePendingJob(ctx, jobID, input.AuthUserID, input.SourceID); err != nil {
		return "", fmt.Errorf("failed to create job")
	}
	if err := s.configureAndStartJob(ctx, jobID, keyColumns, keyColumnDesc, input.ColumnsMetadata, input.MaxRows, len(rowKeys)); err != nil {
		return "", err
	}
	go s.enricher.Enrich(context.Background(), jobID, input.DBUser.ID, stripeCustomerIDOrEmpty(input.DBUser), rowKeys, input.ColumnsMetadata, keyColumnDesc)
	return jobID, nil
}

func (s *sourcesService) configureAndStartJob(ctx context.Context, jobID string, keyColumns []string, keyColumnDesc *string, cols []*models.ColumnMetadata, maxRows *int, rowCount int) error {
	if err := s.store.UpdateJobConfiguration(ctx, jobID, keyColumns, cols, keyColumnDesc, maxRows); err != nil {
		return fmt.Errorf("failed to update job configuration")
	}
	if err := s.store.StartJob(ctx, jobID, rowCount); err != nil {
		return fmt.Errorf("failed to start job")
	}
	return nil
}

func imputationColumnNames(cols []*models.ColumnMetadata) []string {
	var names []string
	for _, col := range cols {
		if col.JobType == models.JobTypeImputation {
			names = append(names, col.Name)
		}
	}
	return names
}

func deduplicateKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			result = append(result, k)
		}
	}
	return result
}

func stripeCustomerIDOrEmpty(u *models.User) string {
	if u.StripeCustomerID != nil {
		return *u.StripeCustomerID
	}
	return ""
}

func generateJobID(extension string) string {
	return uuid.New().String() + extension
}
