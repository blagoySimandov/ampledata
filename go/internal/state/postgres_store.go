package state

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type PostgresStore struct {
	db *bun.DB
}

func NewPostgresStore(connectionString string) (*PostgresStore, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connectionString)))

	db := bun.NewDB(sqldb, pgdialect.New())

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)

	store := &PostgresStore{db: db}

	ctx := context.Background()
	if err := store.InitializeDatabase(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return store, nil
}

func (s *PostgresStore) InitializeDatabase(ctx context.Context) error {
	_, err := s.db.NewCreateTable().
		Model((*models.JobDB)(nil)).
		IfNotExists().
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

	_, err = s.db.NewCreateIndex().
		Model((*models.RowStateDB)(nil)).
		Index("idx_row_states_job_id").
		Column("job_id").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create job_id index: %w", err)
	}

	_, err = s.db.NewCreateIndex().
		Model((*models.RowStateDB)(nil)).
		Index("idx_row_states_stage").
		Column("job_id", "stage").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stage index: %w", err)
	}

	_, err = s.db.NewCreateIndex().
		Model((*models.RowStateDB)(nil)).
		Index("idx_row_states_updated_at").
		Column("updated_at").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create updated_at index: %w", err)
	}

	_, err = s.db.NewCreateIndex().
		Model((*models.JobDB)(nil)).
		Index("idx_jobs_user_id").
		Column("user_id").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user_id index: %w", err)
	}

	_, err = s.db.NewCreateIndex().
		Model((*models.JobDB)(nil)).
		Index("idx_jobs_status").
		Column("status").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
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

func (s *PostgresStore) CreatePendingJob(ctx context.Context, jobID, userID, filePath string) error {
	now := time.Now()
	job := &models.JobDB{
		JobID:     jobID,
		UserID:    userID,
		FilePath:  filePath,
		TotalRows: 0,
		Status:    models.JobStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.NewInsert().
		Model(job).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create pending job: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetJob(ctx context.Context, jobID string) (*models.JobDB, error) {
	var job models.JobDB
	err := s.db.NewSelect().
		Model(&job).
		Where("job_id = ?", jobID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return &job, nil
}

func (s *PostgresStore) UpdateJobConfiguration(ctx context.Context, jobID, keyColumn string, columnsMetadata []*models.ColumnMetadata, entityType *string) error {
	_, err := s.db.NewUpdate().
		Model((*models.JobDB)(nil)).
		Set("key_column = ?", keyColumn).
		Set("columns_metadata = ?", columnsMetadata).
		Set("entity_type = ?", entityType).
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

func (s *PostgresStore) GetJobsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.JobDB, error) {
	var jobs []*models.JobDB
	query := s.db.NewSelect().
		Model(&jobs).
		Where("user_id = ?", userID).
		Order("created_at DESC")

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

		_, err := tx.NewInsert().
			Model(&rows).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert row states: %w", err)
		}

		return nil
	})
}

func (s *PostgresStore) SaveRowState(ctx context.Context, jobID string, state *models.RowState) error {
	dbState := models.RowStateFromApp(jobID, state)

	_, err := s.db.NewInsert().
		Model(dbState).
		On("CONFLICT (job_id, key) DO UPDATE").
		Set("stage = EXCLUDED.stage").
		Set("extracted_data = EXCLUDED.extracted_data").
		Set("confidence = EXCLUDED.confidence").
		Set("error = EXCLUDED.error").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
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

func (s *PostgresStore) GetJobProgress(ctx context.Context, jobID string) (*models.JobProgress, error) {
	var job models.JobDB

	err := s.db.NewSelect().
		Model(&job).
		Where("job_id = ?", jobID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get job info: %w", err)
	}

	type StageCount struct {
		Stage string `bun:"stage"`
		Count int    `bun:"count"`
	}

	var stageCounts []StageCount
	err = s.db.NewSelect().
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

	progress := &models.JobProgress{
		JobID:       jobID,
		TotalRows:   job.TotalRows,
		RowsByStage: rowsByStage,
		Status:      job.Status,
	}
	if job.StartedAt != nil {
		progress.StartedAt = *job.StartedAt
	}
	return progress, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}
