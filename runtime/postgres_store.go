package runtime

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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
	var dsn string
	if c.Provider.PostgreSQL.ConnectionString != "" {
		dsn = c.Provider.PostgreSQL.ConnectionString
	} else {
		// generate dsn
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", c.Provider.PostgreSQL.Host, c.Provider.PostgreSQL.Port, c.Provider.PostgreSQL.User, c.Provider.PostgreSQL.Password, c.Provider.PostgreSQL.Database, c.Provider.PostgreSQL.SSLMode)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return
	}
	// manage the limits of open connections
	if c.Provider.PostgreSQL.MaxLifetimeMs != 0 {
		db.SetConnMaxLifetime(time.Millisecond * time.Duration(c.Provider.PostgreSQL.MaxLifetimeMs))
	}
	if c.Provider.PostgreSQL.MaxIdleTimeMs != 0 {
		db.SetConnMaxIdleTime(time.Millisecond * time.Duration(c.Provider.PostgreSQL.MaxIdleTimeMs))
	}
	if c.Provider.PostgreSQL.MaxOpenConns != 0 {
		db.SetMaxOpenConns(c.Provider.PostgreSQL.MaxOpenConns)
	}
	if c.Provider.PostgreSQL.MaxIdleConns != 0 {
		db.SetMaxIdleConns(c.Provider.PostgreSQL.MaxIdleConns)
	}
	logger.Info("Connected to Postgres")
	pStorage = &PostgresStorage{
		Database: db,
	}
	return
}

// Implementation of the Storage interface methods
func (s *PostgresStorage) ActionSpec(id string) (actionSpec *models.ActionSpec, err error) {
	query := `SELECT * FROM actions WHERE id = $1`
	row := s.Database.QueryRow(query, id)
	actionSpec = &models.ActionSpec{}
	err = row.Scan(&actionSpec.Id, &actionSpec.Name, &actionSpec.Description, &actionSpec.Endpoint)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) ActionSpecs() (actionSpecs []*models.ActionSpec, err error) {
	query := `SELECT * FROM actions`
	rows, err := s.Database.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()
	actionSpecs = make([]*models.ActionSpec, 0)
	for rows.Next() {
		actionSpec := &models.ActionSpec{}
		err = rows.Scan(&actionSpec.Id, &actionSpec.Name, &actionSpec.Description, &actionSpec.Endpoint)
		if err != nil {
			return
		}
		actionSpecs = append(actionSpecs, actionSpec)
	}
	return
}

func (s *PostgresStorage) AddPendingSteps(instanceId string, pendingStep ...*PendingStep) error {
	return nil
}

func (s *PostgresStorage) ActionEndpoint(id string) (endpoint *models.Endpoint, err error) {
	query := `SELECT endpoint FROM actions WHERE id = $1`
	row := s.Database.QueryRow(query, id)
	var endpointJSON []byte
	err = row.Scan(&endpointJSON)
	if err != nil {
		return
	}
	err = json.Unmarshal(endpointJSON, &endpoint)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) ArchiveInstance(workflowID string, archiveInstance bool) error {
	return nil
}

func (s *PostgresStorage) CreateNewInstance(workflowID string, instanceID string, pipeline *data.Pipeline) (err error) {
	query := `INSERT INTO workflow_data (id, workflow_id, pipeline_data) VALUES ($1, $2, $3)`
	mappedData := pipeline.Map()
	pipelineJSON, err := json.Marshal(mappedData)
	if err != nil {
		return
	}
	_, err = s.Database.Exec(query, instanceID, workflowID, pipelineJSON)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) DeleteAction(id string) (err error) {
	query := `UPDATE actions set is_deleted = true WHERE id = $1`
	_, err = s.Database.Exec(query, id)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) DeletePendingStep(instanceID string, pendingStep *PendingStep) error {
	return nil
}

func (s *PostgresStorage) DeleteWorkflow(workflowID string, version int) (err error) {
	query := `Update workflows set is_deleted = true WHERE id = $1 AND version = $2`
	_, err = s.Database.Exec(query, workflowID, version)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) DeleteStepChangeEvent(instanceID, eventID string) error {
	return nil
}

func (s *PostgresStorage) GetPipeline(id string) (pipelineData *data.Pipeline, err error) {
	query := `SELECT pipeline_data FROM workflow_data WHERE id = $1`
	row := s.Database.QueryRow(query, id)
	var pipelineJSON []byte
	err = row.Scan(&pipelineJSON)
	if err != nil {
		return
	}
	var pipelineDataMap map[string]any
	err = json.Unmarshal(pipelineJSON, &pipelineDataMap)
	if err != nil {
		return
	}
	pipelineData = data.NewPipelineFrom(pipelineDataMap)
	return
}

func (s *PostgresStorage) GetState(instanceID string) (state *WorkflowState, err error) {
	query := `SELECT * FROM workflow_state WHERE instance_id = $1`
	row := s.Database.QueryRow(query, instanceID)
	state = &WorkflowState{}
	err = row.Scan(&state.InstanceId, &state.InstanceVersion, &state.WorkflowId, &state.WorkflowVersion, &state.Status, &state.Error)
	if err != nil {
		return
	}
	return
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

func (s *PostgresStorage) GetStepState(instanceID, stepID string) (stepState *StepState, err error) {
	query := `SELECT id, parent_step, child_count, status, instance_id FROM step_state WHERE instance_id = $1 AND step_id = $2`
	row := s.Database.QueryRow(query, instanceID, stepID)
	stepState = &StepState{}
	err = row.Scan(&stepState.StepId, &stepState.ParentStep, &stepState.ChildCount, &stepState.Status, &stepState.InstanceId)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) GetStepStates(instanceID string) (map[string]*StepState, error) {
	return nil, nil
}

func (s *PostgresStorage) GetWorkflow(workflowID string, version int) (workflow *models.Workflow, err error) {
	query := `SELECT * FROM workflows WHERE id = $1 AND version = $2`
	row := s.Database.QueryRow(query, workflowID, version)
	workflow = &models.Workflow{}
	err = row.Scan(&workflow.Id, &workflow.Version, &workflow.Name, &workflow.Description)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) GetWorkflowByInstance(instanceID string) (workflow *models.Workflow, err error) {
	query := `select workflow_id from workflow_data where id = $1`
	row := s.Database.QueryRow(query, instanceID)
	var workflowID string
	err = row.Scan(&workflowID)
	if err != nil {
		return
	}
	query = `select * from workflows where id = $1`
	row = s.Database.QueryRow(query, workflowID)
	workflow = &models.Workflow{}
	err = row.Scan(&workflow.Id, &workflow.Version, &workflow.Name, &workflow.Description)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) ListWorkflows() (workflows []*models.Workflow, err error) {
	query := `SELECT * FROM workflows`
	rows, err := s.Database.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()
	workflows = make([]*models.Workflow, 0)
	for rows.Next() {
		workflow := &models.Workflow{}
		err = rows.Scan(&workflow.Id, &workflow.Version, &workflow.Name, &workflow.Description)
		if err != nil {
			return
		}
		workflows = append(workflows, workflow)
	}
	return
}

func (s *PostgresStorage) ListWorkflowVersions(workflowID string) ([]*models.Workflow, error) {
	return nil, nil
}

func (s *PostgresStorage) ListActions() (actions []*models.ActionSpec, err error) {
	query := `SELECT * FROM actions`
	rows, err := s.Database.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()
	actions = make([]*models.ActionSpec, 0)
	for rows.Next() {
		action := &models.ActionSpec{}
		err = rows.Scan(&action.Id, &action.Name, &action.Description, &action.Endpoint)
		if err != nil {
			return
		}
		actions = append(actions, action)
	}
	return
}

func (s *PostgresStorage) SaveAction(action *models.ActionSpec) (err error) {
	query := `Insert into actions (id, name, description, endpoint) VALUES ($1, $2, $3, $4)`
	endpointJSON, err := json.Marshal(action.Endpoint)
	if err != nil {
		return
	}
	actionId := CreateId()
	_, err = s.Database.Exec(query, actionId, action.Name, action.Description, endpointJSON)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) LockInstance(instanceID string) (bool, error) {
	return false, nil
}

func (s *PostgresStorage) SaveStepChangeEvent(stepEvent *events.StepChangeEvent) error {
	return nil
}

func (s *PostgresStorage) SavePipeline(pipeline *data.Pipeline) (err error) {
	query := `UPDATE workflow_data SET pipeline_data = $1 WHERE id = $2`
	mappedData := pipeline.Map()
	pipelineJSON, err := json.Marshal(mappedData)
	if err != nil {
		return
	}
	_, err = s.Database.Exec(query, pipelineJSON, pipeline.Id())
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) SaveState(workflowState *WorkflowState) (err error) {
	query := `Insert into workflow_state (instance_id, instance_version, workflow_id, workflow_version, status, error) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = s.Database.Exec(query, workflowState.InstanceId, workflowState.InstanceVersion, workflowState.WorkflowId, workflowState.WorkflowVersion, workflowState.Status, workflowState.Error)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) SaveStepState(stepState *StepState) (err error) {
	query := `INSERT INTO step_state (id, parent_step, child_count, status, instance_id) VALUES ($1, $2, $3, $4, $5)`
	_, err = s.Database.Exec(query, stepState.StepId, stepState.ParentStep, stepState.ChildCount, stepState.Status, stepState.InstanceId)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) SaveWorkflow(workflow *models.Workflow) (err error) {
	query := `INSERT INTO workflows (id, version, name, description) VALUES ($1, $2, $3, $4)`
	_, err = s.Database.Exec(query, workflow.Id, workflow.Version, workflow.Name, workflow.Description)
	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) UnlockInstance(instanceID string) (err error) {
	return
}

func (s *PostgresStorage) Config() *config.StorageConfig {
	return &config.StorageConfig{
		Type: config.PostgresStorageType,
		Provider: &config.Provider{
			PostgreSQL: &config.PostgresStorage{},
		},
	}
}
