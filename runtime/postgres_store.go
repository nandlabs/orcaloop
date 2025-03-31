package runtime

import (
	"database/sql"

	_ "github.com/lib/pq"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/events"
	"oss.nandlabs.io/orcaloop-sdk/models"
	"oss.nandlabs.io/orcaloop/config"
)

type PostgresStorage struct {
	Database *sql.DB
}

func ConnectPostgres(c *config.StorageConfig) (pStorage *PostgresStorage, err error) {
	// Connect to Postgres
	db, err := sql.Open("postgres", c.Provider.SQL.ConnectionString)
	if err != nil {
		return
	}
	logger.Info("Connected to Postgres")
	pStorage = &PostgresStorage{
		Database: db,
	}
	return
}

// Implementation of the Storage interface methods
func (s *PostgresStorage) ActionSpec(id string) (*models.ActionSpec, error) {
	return nil, nil
}

func (s *PostgresStorage) ActionSpecs() ([]*models.ActionSpec, error) {
	return nil, nil
}

func (s *PostgresStorage) AddPendingSteps(instanceId string, pendingStep ...*PendingStep) error {
	return nil
}

func (s *PostgresStorage) ActionEndpoint(id string) (*models.Endpoint, error) {
	return nil, nil
}

func (s *PostgresStorage) ArchiveInstance(workflowID string, archiveInstance bool) error {
	return nil
}

func (s *PostgresStorage) CreateNewInstance(workflowID string, instanceID string, pipeline *data.Pipeline) error {
	return nil
}

func (s *PostgresStorage) DeleteAction(id string) error {
	return nil
}

func (s *PostgresStorage) DeletePendingStep(instanceID string, pendingStep *PendingStep) error {
	return nil
}

func (s *PostgresStorage) DeleteWorkflow(workflowID string, version int) error {
	return nil
}

func (s *PostgresStorage) DeleteStepChangeEvent(instanceID, eventID string) error {
	return nil
}

func (s *PostgresStorage) GetPipeline(id string) (*data.Pipeline, error) {
	return nil, nil
}

func (s *PostgresStorage) GetState(instanceID string) (*WorkflowState, error) {
	return nil, nil
}

func (s *PostgresStorage) GetNextPendingStep(instanceID string) (*PendingStep, error) {
	return nil, nil
}

func (s *PostgresStorage) GetPendingSteps(instanceID string) ([]*PendingStep, error) {
	return nil, nil
}

func (s *PostgresStorage) GetStepChangeEvents(instanceID string) ([]*events.StepChangeEvent, error) {
	return nil, nil
}

func (s *PostgresStorage) GetStepState(instanceID, stepID string) (*StepState, error) {
	return nil, nil
}

func (s *PostgresStorage) GetStepStates(instanceID string) (map[string]*StepState, error) {
	return nil, nil
}

func (s *PostgresStorage) GetWorkflow(workflowID string, version int) (*models.Workflow, error) {
	return nil, nil
}

func (s *PostgresStorage) GetWorkflowByInstance(instanceID string) (*models.Workflow, error) {
	return nil, nil
}

func (s *PostgresStorage) ListWorkflows() ([]*models.Workflow, error) {
	return nil, nil
}

func (s *PostgresStorage) ListWorkflowVersions(workflowID string) ([]*models.Workflow, error) {
	return nil, nil
}

func (s *PostgresStorage) ListActions() ([]*models.ActionSpec, error) {
	return nil, nil
}

func (s *PostgresStorage) SaveAction(action *models.ActionSpec) error {
	return nil
}

func (s *PostgresStorage) LockInstance(instanceID string) (bool, error) {
	return false, nil
}

func (s *PostgresStorage) SaveStepChangeEvent(stepEvent *events.StepChangeEvent) error {
	return nil
}

func (s *PostgresStorage) SavePipeline(pipeline *data.Pipeline) error {
	return nil
}

func (s *PostgresStorage) SaveState(workflowState *WorkflowState) error {
	return nil
}

func (s *PostgresStorage) SaveStepState(stepState *StepState) error {
	return nil
}

func (s *PostgresStorage) SaveWorkflow(workflow *models.Workflow) error {
	return nil
}

func (s *PostgresStorage) UnlockInstance(instanceID string) error {
	return nil
}

func (s *PostgresStorage) Config() *config.StorageConfig {
	return nil
}
