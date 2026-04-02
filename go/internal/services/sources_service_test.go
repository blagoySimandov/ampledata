package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/gcs"
	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/blagoySimandov/ampledata/go/internal/services"
	"github.com/blagoySimandov/ampledata/go/internal/state"
	"github.com/google/uuid"
)

// ---- store mock ----

type mockStore struct {
	getSource              func(ctx context.Context, sourceID uuid.UUID) (*models.Source, error)
	getJobsBySource        func(ctx context.Context, sourceID uuid.UUID) ([]*models.Job, error)
	createPendingJob       func(ctx context.Context, jobID, userID string, sourceID uuid.UUID) error
	updateJobConfiguration func(ctx context.Context, jobID string, keyColumns []string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription *string, maxRows *int) error
	startJob               func(ctx context.Context, jobID string, totalRows int) error
}

func (m *mockStore) GetSource(ctx context.Context, id uuid.UUID) (*models.Source, error) {
	return m.getSource(ctx, id)
}
func (m *mockStore) GetJobsBySource(ctx context.Context, id uuid.UUID) ([]*models.Job, error) {
	if m.getJobsBySource != nil {
		return m.getJobsBySource(ctx, id)
	}
	return nil, nil
}
func (m *mockStore) CreatePendingJob(ctx context.Context, jobID, userID string, sourceID uuid.UUID) error {
	if m.createPendingJob != nil {
		return m.createPendingJob(ctx, jobID, userID, sourceID)
	}
	return nil
}
func (m *mockStore) UpdateJobConfiguration(ctx context.Context, jobID string, keyColumns []string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription *string, maxRows *int) error {
	if m.updateJobConfiguration != nil {
		return m.updateJobConfiguration(ctx, jobID, keyColumns, columnsMetadata, keyColumnDescription, maxRows)
	}
	return nil
}
func (m *mockStore) StartJob(ctx context.Context, jobID string, totalRows int) error {
	if m.startJob != nil {
		return m.startJob(ctx, jobID, totalRows)
	}
	return nil
}

// Unused store methods — satisfy the interface.
func (m *mockStore) CreateSource(ctx context.Context, source *models.SourceDB) error { return nil }
func (m *mockStore) GetSourcesByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Source, error) {
	return nil, nil
}
func (m *mockStore) GetJob(ctx context.Context, jobID string) (*models.Job, error) { return nil, nil }
func (m *mockStore) GetJobsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Job, error) {
	return nil, nil
}
func (m *mockStore) BulkCreateRows(ctx context.Context, jobID string, rowKeys []string) error {
	return nil
}
func (m *mockStore) SaveRowState(ctx context.Context, jobID string, s *models.RowState) error {
	return nil
}
func (m *mockStore) GetRowState(ctx context.Context, jobID, key string) (*models.RowState, error) {
	return nil, nil
}
func (m *mockStore) GetRowsAtStage(ctx context.Context, jobID string, stage models.RowStage, offset, limit int) ([]*models.RowState, error) {
	return nil, nil
}
func (m *mockStore) GetRowsPaginated(ctx context.Context, jobID string, params state.RowsQueryParams) (*state.PaginatedRows, error) {
	return nil, nil
}
func (m *mockStore) SetJobStatus(ctx context.Context, jobID string, status models.JobStatus) error {
	return nil
}
func (m *mockStore) GetJobStatus(ctx context.Context, jobID string) (models.JobStatus, error) {
	return "", nil
}
func (m *mockStore) GetJobProgress(ctx context.Context, jobID string) (*models.JobProgress, error) {
	return nil, nil
}
func (m *mockStore) IncrementJobCost(ctx context.Context, jobID string, costDollars, costCredits int) error {
	return nil
}
func (m *mockStore) Close() error { return nil }

// ---- CSV reader mock ----

type mockCSVReader struct {
	readCompositeKey func(ctx context.Context, objectName string, columnNames []string) ([]string, error)
}

func (m *mockCSVReader) ReadCSV(ctx context.Context, objectName string) (*gcs.CSVResult, error) {
	return &gcs.CSVResult{}, nil
}
func (m *mockCSVReader) ReadCompositeKeyFromFile(ctx context.Context, objectName string, columnNames []string) ([]string, error) {
	if m.readCompositeKey != nil {
		return m.readCompositeKey(ctx, objectName, columnNames)
	}
	return nil, nil
}
func (m *mockCSVReader) ReadCompositeKeyFromFileFiltered(ctx context.Context, objectName string, keyColumns []string, filterColumns []string) ([]string, error) {
	if m.readCompositeKey != nil {
		return m.readCompositeKey(ctx, objectName, keyColumns)
	}
	return nil, nil
}
func (m *mockCSVReader) GenerateSignedURL(ctx context.Context, objectName, contentType string) (string, error) {
	return "https://signed-url", nil
}

// ---- enricher mock ----

type mockEnricher struct{}

func (m *mockEnricher) Enrich(ctx context.Context, jobID, userID, stripeCustomerID string, rowKeys []string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription *string) error {
	return nil
}

// ---- helpers ----

// testSource returns a source owned by the given userID.
func testSource(id uuid.UUID, userID string) *models.Source {
	return &models.Source{
		ID:     id,
		UserID: userID,
		Type:   models.SourceTypeCSVUpload,
		Metadata: &models.CSVSourceMetadata{
			FileURI:     "gs://bucket/file.csv",
			ContentType: "text/csv",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// testUser returns a user with enough credits to process the given cell count.
func testUser(tokensIncluded, tokensUsed int64) *models.User {
	db := &models.UserDB{
		ID:             "user-1",
		Email:          "test@example.com",
		TokensIncluded: tokensIncluded,
		TokensUsed:     tokensUsed,
	}
	return db.ToUser()
}

func enrichInput(sourceID uuid.UUID, userID string, user *models.User, keyColumns []string, maxRows *int) services.EnrichSourceInput {
	return services.EnrichSourceInput{
		SourceID:        sourceID,
		AuthUserID:      userID,
		DBUser:          user,
		KeyColumns:      keyColumns,
		ColumnsMetadata: []*models.ColumnMetadata{{Name: "industry", Type: "string", JobType: "enrichment"}},
		MaxRows:         maxRows,
	}
}

// ---- tests ----

func TestEnrichSource_SourceNotFound(t *testing.T) {
	sourceID := uuid.New()
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return nil, errors.New("not found")
		},
	}
	svc := services.NewSourcesService(store, &mockCSVReader{}, &mockEnricher{}, nil, nil)

	_, err := svc.EnrichSource(context.Background(), enrichInput(sourceID, "user-1", testUser(0, 0), []string{"name"}, nil))
	if !errors.Is(err, services.ErrSourceNotFound) {
		t.Errorf("expected ErrSourceNotFound, got %v", err)
	}
}

func TestEnrichSource_SourceForbidden(t *testing.T) {
	sourceID := uuid.New()
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return testSource(sourceID, "other-user"), nil
		},
	}
	svc := services.NewSourcesService(store, &mockCSVReader{}, &mockEnricher{}, nil, nil)

	_, err := svc.EnrichSource(context.Background(), enrichInput(sourceID, "user-1", testUser(0, 0), []string{"name"}, nil))
	if !errors.Is(err, services.ErrSourceForbidden) {
		t.Errorf("expected ErrSourceForbidden, got %v", err)
	}
}

func TestEnrichSource_NoKeyColumnsFirstRun(t *testing.T) {
	sourceID := uuid.New()
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return testSource(sourceID, "user-1"), nil
		},
		getJobsBySource: func(_ context.Context, _ uuid.UUID) ([]*models.Job, error) {
			return []*models.Job{}, nil // no previous jobs
		},
	}
	svc := services.NewSourcesService(store, &mockCSVReader{}, &mockEnricher{}, nil, nil)

	input := enrichInput(sourceID, "user-1", testUser(0, 0), nil /* no key columns */, nil)
	_, err := svc.EnrichSource(context.Background(), input)

	var validErr services.ValidationError
	if !errors.As(err, &validErr) {
		t.Errorf("expected ValidationError, got %v", err)
	}
}

func TestEnrichSource_NoRowsFound(t *testing.T) {
	sourceID := uuid.New()
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return testSource(sourceID, "user-1"), nil
		},
	}
	reader := &mockCSVReader{
		readCompositeKey: func(_ context.Context, _ string, _ []string) ([]string, error) {
			return []string{}, nil // empty CSV
		},
	}
	svc := services.NewSourcesService(store, reader, &mockEnricher{}, nil, nil)

	_, err := svc.EnrichSource(context.Background(), enrichInput(sourceID, "user-1", testUser(0, 0), []string{"name"}, nil))

	var validErr services.ValidationError
	if !errors.As(err, &validErr) {
		t.Errorf("expected ValidationError, got %v", err)
	}
}

func TestEnrichSource_InsufficientCredits(t *testing.T) {
	sourceID := uuid.New()
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return testSource(sourceID, "user-1"), nil
		},
	}
	reader := &mockCSVReader{
		readCompositeKey: func(_ context.Context, _ string, _ []string) ([]string, error) {
			return []string{"row1", "row2", "row3"}, nil
		},
	}
	// user has used all their credits
	user := testUser(0, 100_000)
	svc := services.NewSourcesService(store, reader, &mockEnricher{}, nil, nil)

	_, err := svc.EnrichSource(context.Background(), enrichInput(sourceID, "user-1", user, []string{"name"}, nil))
	if !errors.Is(err, services.ErrInsufficientCredits) {
		t.Errorf("expected ErrInsufficientCredits, got %v", err)
	}
}

func TestEnrichSource_Success(t *testing.T) {
	sourceID := uuid.New()
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return testSource(sourceID, "user-1"), nil
		},
	}
	reader := &mockCSVReader{
		readCompositeKey: func(_ context.Context, _ string, _ []string) ([]string, error) {
			return []string{"row1", "row2"}, nil
		},
	}
	svc := services.NewSourcesService(store, reader, &mockEnricher{}, nil, nil)

	jobID, err := svc.EnrichSource(context.Background(), enrichInput(sourceID, "user-1", testUser(0, 0), []string{"name"}, nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jobID == "" {
		t.Error("expected non-empty jobID")
	}
}

func TestEnrichSource_MaxRowsLimitsRows(t *testing.T) {
	sourceID := uuid.New()

	var capturedRowsCount int
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return testSource(sourceID, "user-1"), nil
		},
		startJob: func(_ context.Context, _ string, totalRows int) error {
			capturedRowsCount = totalRows
			return nil
		},
	}
	reader := &mockCSVReader{
		readCompositeKey: func(_ context.Context, _ string, _ []string) ([]string, error) {
			// Return 10 rows, but maxRows will cap at 3.
			return []string{"r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "r10"}, nil
		},
	}
	svc := services.NewSourcesService(store, reader, &mockEnricher{}, nil, nil)

	maxRows := 3
	_, err := svc.EnrichSource(context.Background(), enrichInput(sourceID, "user-1", testUser(0, 0), []string{"name"}, &maxRows))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedRowsCount != maxRows {
		t.Errorf("StartJob called with %d rows, want %d", capturedRowsCount, maxRows)
	}
}

func TestEnrichSource_StoreErrorOnCreatePendingJob(t *testing.T) {
	sourceID := uuid.New()
	storeErr := errors.New("db connection lost")
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return testSource(sourceID, "user-1"), nil
		},
		createPendingJob: func(_ context.Context, _ string, _ string, _ uuid.UUID) error {
			return storeErr
		},
	}
	reader := &mockCSVReader{
		readCompositeKey: func(_ context.Context, _ string, _ []string) ([]string, error) {
			return []string{"row1"}, nil
		},
	}
	svc := services.NewSourcesService(store, reader, &mockEnricher{}, nil, nil)

	_, err := svc.EnrichSource(context.Background(), enrichInput(sourceID, "user-1", testUser(0, 0), []string{"name"}, nil))
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

func TestEnrichSource_KeyColumnsInheritedFromPreviousJob(t *testing.T) {
	sourceID := uuid.New()
	prevKeyColumns := []string{"company_name"}
	store := &mockStore{
		getSource: func(_ context.Context, _ uuid.UUID) (*models.Source, error) {
			return testSource(sourceID, "user-1"), nil
		},
		getJobsBySource: func(_ context.Context, _ uuid.UUID) ([]*models.Job, error) {
			return []*models.Job{{KeyColumns: prevKeyColumns}}, nil
		},
	}
	reader := &mockCSVReader{
		readCompositeKey: func(_ context.Context, _ string, cols []string) ([]string, error) {
			if len(cols) == 0 || cols[0] != prevKeyColumns[0] {
				return nil, errors.New("unexpected key columns")
			}
			return []string{"Acme Corp"}, nil
		},
	}
	svc := services.NewSourcesService(store, reader, &mockEnricher{}, nil, nil)

	// submit with no key columns — should inherit from previous job
	input := enrichInput(sourceID, "user-1", testUser(0, 0), nil, nil)
	jobID, err := svc.EnrichSource(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jobID == "" {
		t.Error("expected non-empty jobID")
	}
}
