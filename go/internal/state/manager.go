package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
	"github.com/google/uuid"
)

// TODO: use IStateManager interface
type StateManager struct {
	store       Store
	cancelFuncs map[string]context.CancelFunc
	workflowIDs map[string]workflowIDPair // jobID -> workflow ID and run ID
	mu          sync.RWMutex
}

type workflowIDPair struct {
	WorkflowID string
	RunID      string
}

func NewStateManager(store Store) *StateManager {
	return &StateManager{
		store:       store,
		cancelFuncs: make(map[string]context.CancelFunc),
		workflowIDs: make(map[string]workflowIDPair),
	}
}

func (m *StateManager) GenerateJobID() string {
	return uuid.New().String()
}

func (m *StateManager) InitializeJob(ctx context.Context, jobID string, rowKeys []string) error {
	job, err := m.store.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	if job.Status != models.JobStatusRunning {
		return fmt.Errorf("job %s is not in RUNNING status, current status: %s", jobID, job.Status)
	}

	if err := m.store.BulkCreateRows(ctx, jobID, rowKeys); err != nil {
		return fmt.Errorf("failed to create rows: %w", err)
	}

	return nil
}

func (m *StateManager) RegisterCancelFunc(jobID string, cancel context.CancelFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancelFuncs[jobID] = cancel
}

func (m *StateManager) RegisterWorkflowID(jobID, workflowID, runID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflowIDs[jobID] = workflowIDPair{
		WorkflowID: workflowID,
		RunID:      runID,
	}
}

func (m *StateManager) GetWorkflowID(jobID string) (string, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if pair, ok := m.workflowIDs[jobID]; ok {
		return pair.WorkflowID, pair.RunID
	}
	return "", ""
}

func (m *StateManager) Transition(ctx context.Context, jobID, key string, toStage models.RowStage, dataUpdate map[string]interface{}) error {
	cancelled, err := m.CheckCancelled(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to check cancellation: %w", err)
	}

	if cancelled {
		state, err := m.store.GetRowState(ctx, jobID, key)
		if err == nil && state != nil {
			if state.Stage != models.StageCompleted && state.Stage != models.StageFailed {
				state.Stage = models.StageCancelled
				state.UpdatedAt = time.Now()
				m.store.SaveRowState(ctx, jobID, state)
			}
		}
		return fmt.Errorf("job %s was cancelled", jobID)
	}

	state, err := m.store.GetRowState(ctx, jobID, key)
	if err != nil {
		return fmt.Errorf("failed to get row state: %w", err)
	}

	if state == nil {
		return fmt.Errorf("no state found for key %s", key)
	}

	state.Stage = toStage
	state.UpdatedAt = time.Now()

	if dataUpdate != nil {
		if extractedData, ok := dataUpdate["extracted_data"]; ok {
			if data, ok := extractedData.(map[string]interface{}); ok {
				state.ExtractedData = data
			}
		}
		if confidence, ok := dataUpdate["confidence"]; ok {
			if data, ok := confidence.(map[string]*models.FieldConfidenceInfo); ok {
				state.Confidence = data
			}
		}
		if errMsg, ok := dataUpdate["error"]; ok {
			if errStr, ok := errMsg.(string); ok {
				state.Error = &errStr
				state.Stage = models.StageFailed
			}
		}
	}

	if err := m.store.SaveRowState(ctx, jobID, state); err != nil {
		return fmt.Errorf("failed to save row state: %w", err)
	}

	return nil
}

func (m *StateManager) GetPendingForStage(ctx context.Context, jobID string, stage models.RowStage) ([]*models.RowState, error) {
	prerequisiteMap := map[models.RowStage]models.RowStage{
		models.StageSerpFetched:  models.StagePending,
		models.StageDecisionMade: models.StageSerpFetched,
		models.StageCrawled:      models.StageDecisionMade,
		models.StageEnriched:     models.StageCrawled,
		models.StageCompleted:    models.StageEnriched,
	}

	requiredStage, ok := prerequisiteMap[stage]
	if !ok {
		requiredStage = models.StagePending
	}

	return m.store.GetRowsAtStage(ctx, jobID, requiredStage, 0, 0)
}

func (m *StateManager) CheckCancelled(ctx context.Context, jobID string) (bool, error) {
	m.mu.RLock()
	_, hasCancelFunc := m.cancelFuncs[jobID]
	m.mu.RUnlock()

	if hasCancelFunc {
		select {
		case <-ctx.Done():
			return true, nil
		default:
		}
	}

	status, err := m.store.GetJobStatus(ctx, jobID)
	if err != nil {
		return false, err
	}

	return status == models.JobStatusPaused ||
		status == models.JobStatusCancelled ||
		status == models.JobStatusCompleted, nil
}

func (m *StateManager) Cancel(ctx context.Context, jobID string) error {
	m.mu.Lock()
	if cancel, ok := m.cancelFuncs[jobID]; ok {
		cancel()
	}
	m.mu.Unlock()

	return m.store.SetJobStatus(ctx, jobID, models.JobStatusCancelled)
}

func (m *StateManager) Pause(ctx context.Context, jobID string) error {
	m.mu.Lock()
	if cancel, ok := m.cancelFuncs[jobID]; ok {
		cancel()
	}
	m.mu.Unlock()

	return m.store.SetJobStatus(ctx, jobID, models.JobStatusPaused)
}

func (m *StateManager) Resume(ctx context.Context, jobID string) error {
	return m.store.SetJobStatus(ctx, jobID, models.JobStatusRunning)
}

func (m *StateManager) Complete(ctx context.Context, jobID string) error {
	return m.store.SetJobStatus(ctx, jobID, models.JobStatusCompleted)
}

func (m *StateManager) Progress(ctx context.Context, jobID string) (*models.JobProgress, error) {
	return m.store.GetJobProgress(ctx, jobID)
}

func (m *StateManager) Store() Store {
	return m.store
}
