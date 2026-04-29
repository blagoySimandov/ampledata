package state

import (
	"context"
	"fmt"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PostgresStore struct {
	db *bun.DB
}

func NewPostgresStore(db *bun.DB) (*PostgresStore, error) {
	store := &PostgresStore{db: db}

	ctx := context.Background()
	if err := store.InitializeDatabase(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return store, nil
}

func (s *PostgresStore) InitializeDatabase(ctx context.Context) error {
	if err := s.createEnumTypes(ctx); err != nil {
		return err
	}
	if err := s.createTables(ctx); err != nil {
		return err
	}
	return s.createIndexes(ctx)
}

func (s *PostgresStore) createEnumTypes(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		DO $$ BEGIN
			CREATE TYPE source_type AS ENUM ('csv_upload');
		EXCEPTION WHEN duplicate_object THEN null;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to create enum types: %w", err)
	}
	return nil
}

func (s *PostgresStore) createTables(ctx context.Context) error {
	_, err := s.db.NewCreateTable().
		Model((*models.UserDB)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	_, err = s.db.NewCreateTable().
		Model((*models.SourceDB)(nil)).
		IfNotExists().
		ForeignKey(`("user_id") REFERENCES "users" ("id")`).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create sources table: %w", err)
	}

	_, err = s.db.NewCreateTable().
		Model((*models.JobDB)(nil)).
		IfNotExists().
		ForeignKey(`("source_id") REFERENCES "sources" ("id")`).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create jobs table: %w", err)
	}

	_, err = s.db.NewCreateTable().
		Model((*models.RowStateDB)(nil)).
		IfNotExists().
		ForeignKey(`("job_id") REFERENCES "jobs" ("job_id") ON DELETE CASCADE`).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create row_states table: %w", err)
	}

	return nil
}

func (s *PostgresStore) createIndexes(ctx context.Context) error {
	indexes := []struct {
		model  any
		name   string
		column []string
	}{
		{(*models.SourceDB)(nil), "idx_sources_user_id", []string{"user_id"}},
		{(*models.JobDB)(nil), "idx_jobs_user_id", []string{"user_id"}},
		{(*models.JobDB)(nil), "idx_jobs_status", []string{"status"}},
		{(*models.RowStateDB)(nil), "idx_row_states_job_id", []string{"job_id"}},
		{(*models.RowStateDB)(nil), "idx_row_states_stage", []string{"job_id", "stage"}},
		{(*models.RowStateDB)(nil), "idx_row_states_updated_at", []string{"updated_at"}},
	}

	for _, idx := range indexes {
		_, err := s.db.NewCreateIndex().
			Model(idx.model).
			Index(idx.name).
			Column(idx.column...).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}

	return nil
}

func (s *PostgresStore) CreateJob(ctx context.Context, jobID string, totalRows int, status models.JobStatus) error {
	now := time.Now()
	job := &models.JobDB{
		JobID:     jobID,
		UserID:    "",
		TotalRows: totalRows,
		Status:    status,
		StartedAt: &now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.NewInsert().
		Model(job).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	return nil
}

func (s *PostgresStore) CreateSource(ctx context.Context, source *models.SourceDB) error {
	_, err := s.db.NewInsert().Model(source).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetSource(ctx context.Context, sourceID uuid.UUID) (*models.Source, error) {
	var sourceDB models.SourceDB
	err := s.db.NewSelect().Model(&sourceDB).Where("id = ?", sourceID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	return sourceDB.ToSource()
}

func (s *PostgresStore) GetSourcesByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Source, error) {
	var sourcesDB []*models.SourceDB
	query := s.db.NewSelect().Model(&sourcesDB).Where("src.user_id = ?", userID).Order("src.created_at DESC")
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}
	sources := make([]*models.Source, len(sourcesDB))
	for i, src := range sourcesDB {
		s, err := src.ToSource()
		if err != nil {
			return nil, err
		}
		sources[i] = s
	}
	return sources, nil
}

func (s *PostgresStore) GetJobsBySource(ctx context.Context, sourceID uuid.UUID) ([]*models.Job, error) {
	var jobsDB []*models.JobDB
	err := s.db.NewSelect().Model(&jobsDB).Where("j.source_id = ?", sourceID).Order("j.created_at DESC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by source: %w", err)
	}
	jobs := make([]*models.Job, len(jobsDB))
	for i, jobDB := range jobsDB {
		job, err := jobDB.ToJob()
		if err != nil {
			return nil, err
		}
		jobs[i] = job
	}
	return jobs, nil
}

func (s *PostgresStore) CreatePendingJob(ctx context.Context, jobID, userID string, sourceID uuid.UUID, templateID *uuid.UUID) error {
	now := time.Now()
	job := &models.JobDB{
		JobID:      jobID,
		UserID:     userID,
		SourceID:   sourceID,
		TemplateID: templateID,
		TotalRows:  0,
		Status:     models.JobStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err := s.db.NewInsert().Model(job).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create pending job: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	var jobDB models.JobDB
	err := s.db.NewSelect().
		Model(&jobDB).
		Relation("Source").
		Where("j.job_id = ?", jobID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return jobDB.ToJob()
}

func (s *PostgresStore) UpdateJobConfiguration(ctx context.Context, jobID string, keyColumns []string, columnsMetadata []*models.ColumnMetadata, keyColumnDescription *string) error {
	_, err := s.db.NewUpdate().
		Model((*models.JobDB)(nil)).
		Set("key_columns = ?", keyColumns).
		Set("columns_metadata = ?", columnsMetadata).
		Set("entity_type = ?", keyColumnDescription).
		Set("updated_at = ?", time.Now()).
		Where("job_id = ?", jobID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update job configuration: %w", err)
	}
	return nil
}

func (s *PostgresStore) StartJob(ctx context.Context, jobID string, totalRows int) error {
	now := time.Now()
	_, err := s.db.NewUpdate().
		Model((*models.JobDB)(nil)).
		Set("status = ?", models.JobStatusRunning).
		Set("total_rows = ?", totalRows).
		Set("started_at = ?", now).
		Set("updated_at = ?", now).
		Where("job_id = ?", jobID).
		Where("status = ?", models.JobStatusPending).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to start job: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetJobsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.Job, error) {
	var jobsDB []*models.JobDB
	query := s.db.NewSelect().
		Model(&jobsDB).
		Relation("Source").
		Where("j.user_id = ?", userID).
		Order("j.created_at DESC")

	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by user: %w", err)
	}

	jobs := make([]*models.Job, len(jobsDB))
	for i, jobDB := range jobsDB {
		job, err := jobDB.ToJob()
		if err != nil {
			return nil, fmt.Errorf("failed to convert job: %w", err)
		}
		jobs[i] = job
	}
	return jobs, nil
}

func (s *PostgresStore) BulkCreateRows(ctx context.Context, jobID string, rowKeys []string) error {
	return s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		now := time.Now()
		rows := make([]models.RowStateDB, len(rowKeys))

		for i, key := range rowKeys {
			rows[i] = models.RowStateDB{
				JobID:     jobID,
				Key:       key,
				Stage:     models.StagePending,
				CreatedAt: now,
				UpdatedAt: now,
			}
		}

		// Only insert non-JSONB columns to avoid storing JSON "null"
		// JSONB columns will be database NULL by default
		// DO NOTHING on conflict makes this idempotent: Temporal may retry
		// InitializeJob after a transient failure, and rows inserted in a
		// previous attempt must not cause the retry to fail.
		_, err := tx.NewInsert().
			Model(&rows).
			Column("job_id", "key", "stage", "created_at", "updated_at").
			On("CONFLICT (job_id, key) DO NOTHING").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert row states: %w", err)
		}

		return nil
	})
}

func prepareDBState(jobID string, state *models.RowState) *models.RowStateDB {
	dbState := models.RowStateFromApp(jobID, state)
	now := time.Now()
	dbState.UpdatedAt = now
	if dbState.CreatedAt.IsZero() {
		dbState.CreatedAt = now
	}
	return dbState
}

func determineColumns(state *models.RowState) (insertCols, updateCols []string) {
	insertCols = []string{"job_id", "key", "stage", "created_at", "updated_at"}
	updateCols = []string{"stage", "updated_at"}

	if state.ExtractedData != nil {
		insertCols = append(insertCols, "extracted_data")
		updateCols = append(updateCols, "extracted_data")
	}
	if state.Confidence != nil {
		insertCols = append(insertCols, "confidence")
		updateCols = append(updateCols, "confidence")
	}
	if state.Sources != nil {
		insertCols = append(insertCols, "sources")
		updateCols = append(updateCols, "sources")
	}
	if state.ExtractionHistory != nil {
		insertCols = append(insertCols, "extraction_history")
		updateCols = append(updateCols, "extraction_history")
	}
	if state.Error != nil {
		insertCols = append(insertCols, "error")
		updateCols = append(updateCols, "error")
	}
	return
}

func (s *PostgresStore) SaveRowState(ctx context.Context, jobID string, state *models.RowState) error {
	dbState := prepareDBState(jobID, state)
	insertCols, updateCols := determineColumns(state)

	query := s.db.NewInsert().
		Model(dbState).
		Column(insertCols...).
		On("CONFLICT (job_id, key) DO UPDATE")

	for _, col := range updateCols {
		query = query.Set("? = EXCLUDED.?", bun.Ident(col), bun.Ident(col))
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save row state: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetRowState(ctx context.Context, jobID string, key string) (*models.RowState, error) {
	var dbState models.RowStateDB

	err := s.db.NewSelect().
		Model(&dbState).
		Where("job_id = ?", jobID).
		Where("key = ?", key).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get row state: %w", err)
	}

	return dbState.ToRowState(), nil
}

func (s *PostgresStore) GetRowsAtStage(ctx context.Context, jobID string, stage models.RowStage, offset, limit int) ([]*models.RowState, error) {
	var dbStates []models.RowStateDB

	query := s.db.NewSelect().
		Model(&dbStates).
		Where("job_id = ?", jobID).
		Where("stage = ?", stage).
		Order("created_at ASC")

	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows at stage: %w", err)
	}

	states := make([]*models.RowState, len(dbStates))
	for i := range dbStates {
		states[i] = dbStates[i].ToRowState()
	}

	return states, nil
}

func (s *PostgresStore) BulkCancelActiveRows(ctx context.Context, jobID string) error {
	_, err := s.db.NewUpdate().
		Model((*models.RowStateDB)(nil)).
		Set("stage = ?", models.StageCancelled).
		Set("updated_at = ?", time.Now()).
		Where("job_id = ?", jobID).
		Where("stage NOT IN (?)", bun.In([]models.RowStage{
			models.StageCompleted,
			models.StageFailed,
			models.StageCancelled,
		})).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to bulk cancel active rows: %w", err)
	}
	return nil
}

func (s *PostgresStore) SetJobStatus(ctx context.Context, jobID string, status models.JobStatus) error {
	_, err := s.db.NewUpdate().
		Model((*models.JobDB)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", time.Now()).
		Where("job_id = ?", jobID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set job status: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetJobStatus(ctx context.Context, jobID string) (models.JobStatus, error) {
	var job models.JobDB

	err := s.db.NewSelect().
		Model(&job).
		Column("status").
		Where("job_id = ?", jobID).
		Scan(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get job status: %w", err)
	}

	return job.Status, nil
}

type stageCount struct {
	Stage string `bun:"stage"`
	Count int    `bun:"count"`
}

func (s *PostgresStore) fetchStageCounts(ctx context.Context, jobID string) (map[models.RowStage]int, error) {
	var stageCounts []stageCount
	err := s.db.NewSelect().
		Model((*models.RowStateDB)(nil)).
		Column("stage").
		ColumnExpr("COUNT(*) as count").
		Where("job_id = ?", jobID).
		Group("stage").
		Scan(ctx, &stageCounts)
	if err != nil {
		return nil, fmt.Errorf("failed to get row counts: %w", err)
	}

	rowsByStage := make(map[models.RowStage]int)
	for _, sc := range stageCounts {
		rowsByStage[models.RowStage(sc.Stage)] = sc.Count
	}
	return rowsByStage, nil
}

func buildJobProgress(jobID string, job *models.JobDB, rowsByStage map[models.RowStage]int) *models.JobProgress {
	progress := &models.JobProgress{
		JobID:       jobID,
		TotalRows:   job.TotalRows,
		RowsByStage: rowsByStage,
		Status:      job.Status,
		CostDollars: job.CostDollars,
		CostCredits: job.CostCredits,
	}
	if job.StartedAt != nil {
		progress.StartedAt = *job.StartedAt
	}
	return progress
}

func (s *PostgresStore) GetJobProgress(ctx context.Context, jobID string) (*models.JobProgress, error) {
	var job models.JobDB
	err := s.db.NewSelect().
		Model(&job).
		Where("job_id = ?", jobID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get job info: %w", err)
	}

	rowsByStage, err := s.fetchStageCounts(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return buildJobProgress(jobID, &job, rowsByStage), nil
}

func (s *PostgresStore) IncrementJobCost(ctx context.Context, jobID string, costDollars, costCredits int) error {
	_, err := s.db.NewUpdate().
		Model((*models.JobDB)(nil)).
		Set("cost_dollars = cost_dollars + ?", costDollars).
		Set("cost_credits = cost_credits + ?", costCredits).
		Set("updated_at = ?", time.Now()).
		Where("job_id = ?", jobID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to increment job cost: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetRowsPaginated(ctx context.Context, jobID string, params RowsQueryParams) (*PaginatedRows, error) {
	var dbStates []models.RowStateDB
	var total int
	var err error

	countQuery := s.db.NewSelect().
		Model((*models.RowStateDB)(nil)).
		Where("job_id = ?", jobID)

	if params.Stage != "" && params.Stage != "all" {
		countQuery = countQuery.Where("stage = ?", params.Stage)
	}

	total, err = countQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count rows: %w", err)
	}

	query := s.db.NewSelect().
		Model(&dbStates).
		Where("job_id = ?", jobID)

	if params.Stage != "" && params.Stage != "all" {
		query = query.Where("stage = ?", params.Stage)
	}

	switch params.Sort {
	case "updated_at_asc":
		query = query.Order("updated_at ASC")
	case "key_asc":
		query = query.Order("key ASC")
	case "key_desc":
		query = query.Order("key DESC")
	default:
		query = query.Order("updated_at DESC")
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}

	states := make([]*models.RowState, len(dbStates))
	for i := range dbStates {
		states[i] = dbStates[i].ToRowState()
	}

	return &PaginatedRows{
		Rows:  states,
		Total: total,
	}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}
