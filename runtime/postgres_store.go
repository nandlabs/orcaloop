package runtime

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	query := `SELECT * FROM actions WHERE id = ? AND is_deleted = false`
	sqlStatement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row := sqlStatement.QueryRow(id)
	actionSpec = &models.ActionSpec{}
	err = row.Scan(&actionSpec.Id, &actionSpec.Name, &actionSpec.Description, &actionSpec.Endpoint)
	if err != nil {
		logger.Error("Action not found: %v", err)
		err = errors.New("action not found")
		return
	}
	return
}

func (s *PostgresStorage) ActionSpecs() (actionSpecs []*models.ActionSpec, err error) {
	query := `SELECT * FROM actions WHERE is_deleted = false`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	rows, err := statement.Query(query)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error fetching action specs")
		return
	}
	defer rows.Close()
	actionSpecs = make([]*models.ActionSpec, 0)
	for rows.Next() {
		actionSpec := &models.ActionSpec{}
		err = rows.Scan(&actionSpec.Id, &actionSpec.Name, &actionSpec.Description, &actionSpec.Endpoint)
		if err != nil {
			logger.ErrorF("Error scanning row: %v", err)
			err = errors.New("error scanning action spec row")
			return
		}
		actionSpecs = append(actionSpecs, actionSpec)
	}
	return
}

func (s *PostgresStorage) AddPendingSteps(instanceId string, pendingStep ...*PendingStep) (err error) {
	// serialize this pendingStep and save it to data
	query := `INSERT INTO pending_steps (instance_id, step_id, data) VALUES (?, ?, ?)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	for _, step := range pendingStep {
		var pendingStepJSON []byte
		pendingStepJSON, err = json.Marshal(step)
		if err != nil {
			logger.ErrorF("Error marshalling pending step: %v", err)
			err = errors.New("error marshalling pending step")
			return
		}
		_, err = statement.Exec(query, instanceId, step.StepId, pendingStepJSON)
		if err != nil {
			logger.ErrorF("Error executing query: %v", err)
			err = errors.New("error adding pending step/s")
			return
		}
	}
	return
}

func (s *PostgresStorage) ActionEndpoint(id string) (endpoint *models.Endpoint, err error) {
	query := `SELECT endpoint FROM actions WHERE id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row := statement.QueryRow(query, id)
	var endpointJSON []byte
	err = row.Scan(&endpointJSON)
	if err != nil {
		logger.ErrorF("Error scanning row: %v", err)
		err = errors.New("error scanning action endpoint")
		return
	}
	err = json.Unmarshal(endpointJSON, &endpoint)
	if err != nil {
		logger.ErrorF("Error unmarshalling endpoint: %v", err)
		err = errors.New("error unmarshalling action endpoint")
		return
	}
	return
}

func (s *PostgresStorage) ArchiveInstance(workflowID string, archiveInstance bool) error {
	// to be implemented
	return nil
}

func (s *PostgresStorage) CreateNewInstance(workflowID string, instanceID string, pipeline *data.Pipeline) (err error) {
	query := `INSERT INTO workflow_data (instance_id, workflow_id, pipeline_data) VALUES (?, ?, ?)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	mappedData := pipeline.Map()
	pipelineJSON, err := json.Marshal(mappedData)
	if err != nil {
		logger.ErrorF("Error marshalling pipeline data: %v", err)
		err = errors.New("error marshalling pipeline data")
		return
	}
	_, err = statement.Exec(query, instanceID, workflowID, pipelineJSON)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error creating new instance")
		return
	}
	return
}

func (s *PostgresStorage) DeleteAction(id string) (err error) {
	query := `UPDATE actions set is_deleted = true WHERE id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, id)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error deleting action")
		return
	}
	return
}

func (s *PostgresStorage) DeletePendingStep(instanceID string, pendingStep *PendingStep) (err error) {
	// soft delete based on step_id and instance_id
	query := `UPDATE pending_steps SET is_deleted = true WHERE instance_id = ? AND step_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, instanceID, pendingStep.StepId)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error deleting pending step")
		return
	}
	return nil
}

func (s *PostgresStorage) DeleteWorkflow(workflowID string, version int) (err error) {
	query := `Update workflows set is_deleted = true WHERE workflow_id = ? AND version = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, workflowID, version)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error deleting workflow")
		return
	}
	return
}

func (s *PostgresStorage) DeleteStepChangeEvent(instanceID, eventID string) (err error) {
	// soft delete
	query := `UPDATE step_change_event SET is_deleted = true WHERE instance_id = ? AND event_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, instanceID, eventID)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error deleting step change event")
		return
	}
	return nil
}

func (s *PostgresStorage) GetPipeline(id string) (pipelineData *data.Pipeline, err error) {
	query := `SELECT pipeline_data FROM workflow_data WHERE instance_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row := statement.QueryRow(query, id)
	var pipelineJSON []byte
	err = row.Scan(&pipelineJSON)
	if err != nil {
		logger.ErrorF("Error scanning row: %v", err)
		err = errors.New("error scanning pipeline data")
		return
	}
	var pipelineDataMap map[string]any
	err = json.Unmarshal(pipelineJSON, &pipelineDataMap)
	if err != nil {
		logger.ErrorF("Error unmarshalling pipeline data: %v", err)
		err = errors.New("error unmarshalling pipeline data")
		return
	}
	pipelineData = data.NewPipelineFrom(pipelineDataMap)
	return
}

func (s *PostgresStorage) GetState(instanceID string) (state *WorkflowState, err error) {
	query := `SELECT * FROM workflow_state WHERE instance_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row := statement.QueryRow(query, instanceID)
	state = &WorkflowState{}
	err = row.Scan(&state.InstanceId, &state.InstanceVersion, &state.WorkflowId, &state.WorkflowVersion, &state.Status, &state.Error)
	if err != nil {
		logger.ErrorF("Error scanning row: %v", err)
		err = errors.New("error scanning workflow state")
		return
	}
	return
}

func (s *PostgresStorage) GetNextPendingStep(instanceID string) (pendingStep *PendingStep, err error) {
	// fetch the earliest pending step based on the created timestamp
	query := `SELECT data FROM pending_steps WHERE instance_id = ? AND is_deleted = false ORDER BY created_at ASC LIMIT 1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row := statement.QueryRow(query, instanceID)
	pendingStep = &PendingStep{}
	var pendingStepJSON []byte
	err = row.Scan(&pendingStepJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// no pending steps found
			return nil, nil
		}
		logger.ErrorF("Error scanning row: %v", err)
		err = errors.New("error scanning pending step")
		return
	}
	err = json.Unmarshal(pendingStepJSON, pendingStep)
	if err != nil {
		logger.ErrorF("Error unmarshalling pending step: %v", err)
		err = errors.New("error unmarshalling pending step")
		return
	}
	return
}

func (s *PostgresStorage) GetPendingSteps(instanceID string) (pendingSteps []*PendingStep, err error) {
	query := `SELECT data FROM pending_steps WHERE instance_id = ? AND is_deleted = false`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	rows, err := statement.Query(query, instanceID)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error fetching pending steps")
		return
	}
	defer rows.Close()
	pendingSteps = make([]*PendingStep, 0)
	for rows.Next() {
		pendingStep := &PendingStep{}
		var pendingStepJSON []byte
		err = rows.Scan(&pendingStepJSON)
		if err != nil {
			logger.ErrorF("Error scanning row: %v", err)
			err = errors.New("error scanning pending step")
			return
		}
		err = json.Unmarshal(pendingStepJSON, pendingStep)
		if err != nil {
			logger.ErrorF("Error unmarshalling pending step: %v", err)
			err = errors.New("error unmarshalling pending step")
			return
		}
		pendingSteps = append(pendingSteps, pendingStep)
	}
	return
}

func (s *PostgresStorage) GetStepChangeEvents(instanceID string) (stepChangeEvents []*events.StepChangeEvent, err error) {
	query := `SELECT instance_id, event_id, status, data  FROM step_change_event WHERE instance_id = ? AND is_deleted = false`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	rows, err := statement.Query(query, instanceID)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error fetching step change events")
		return
	}
	defer rows.Close()
	stepChangeEvents = make([]*events.StepChangeEvent, 0)
	for rows.Next() {
		stepChangeEvent := &events.StepChangeEvent{}
		var stepChangeEventDataJSON []byte
		err = rows.Scan(&stepChangeEvent.InstanceId, &stepChangeEvent.EventId, &stepChangeEvent.Status, &stepChangeEventDataJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				// no step change events found
				return
			}
			logger.ErrorF("Error scanning row: %v", err)
			err = errors.New("error scanning step change event")
			return
		}
		err = json.Unmarshal(stepChangeEventDataJSON, &stepChangeEvent.Data)
		if err != nil {
			logger.ErrorF("Error unmarshalling step change event data: %v", err)
			err = errors.New("error unmarshalling step change event data")
			return
		}
		stepChangeEvents = append(stepChangeEvents, stepChangeEvent)
	}
	return
}

func (s *PostgresStorage) GetStepState(instanceID, stepID string) (stepState *StepState, err error) {
	query := `SELECT step_id, parent_step, child_count, status, instance_id FROM step_state WHERE instance_id = ? AND step_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row := statement.QueryRow(query, instanceID, stepID)
	stepState = &StepState{}
	err = row.Scan(&stepState.StepId, &stepState.ParentStep, &stepState.ChildCount, &stepState.Status, &stepState.InstanceId)
	if err != nil {
		if err == sql.ErrNoRows {
			// no step state found
			return nil, nil
		}
		logger.ErrorF("Error scanning row: %v", err)
		err = errors.New("error scanning step state")
		return
	}
	return
}

func (s *PostgresStorage) GetStepStates(instanceID string) (stepState map[string]*StepState, err error) {
	query := `SELECT step_state FROM step_state WHERE instance_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	rows, err := statement.Query(query, instanceID)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error fetching step states")
		return
	}
	defer rows.Close()
	stepState = make(map[string]*StepState)
	for rows.Next() {
		stepStateItem := &StepState{}
		var stepStateJSON []byte
		err = rows.Scan(&stepStateJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				// no step state found
				return nil, nil
			}
			logger.ErrorF("Error scanning row: %v", err)
			err = errors.New("error scanning step state")
			return
		}
		err = json.Unmarshal(stepStateJSON, stepStateItem)
		if err != nil {
			logger.ErrorF("Error unmarshalling step state: %v", err)
			err = errors.New("error unmarshalling step state")
			return
		}
		stepState[stepStateItem.StepId] = stepStateItem
	}
	return
}

func (s *PostgresStorage) GetWorkflow(workflowID string, version int) (workflow *models.Workflow, err error) {
	query := `SELECT * FROM workflows WHERE workflow_id = ? AND version = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row := statement.QueryRow(query, workflowID, version)
	workflow = &models.Workflow{}
	err = row.Scan(&workflow.Id, &workflow.Version, &workflow.Name, &workflow.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			// no workflow found
			return nil, nil
		}
		logger.ErrorF("Error scanning row: %v", err)
		err = errors.New("error scanning workflow")
		return
	}
	return
}

func (s *PostgresStorage) GetWorkflowByInstance(instanceID string) (workflow *models.Workflow, err error) {
	query := `select workflow_id from workflow_data where id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row := statement.QueryRow(query, instanceID)
	var workflowID string
	err = row.Scan(&workflowID)
	if err != nil {
		if err == sql.ErrNoRows {
			// no workflow found
			return nil, nil
		}
		logger.ErrorF("Error scanning row: %v", err)
		err = errors.New("error scanning workflow")
		return
	}
	query = `select * from workflows where workflow_id = ?`
	statement, err = s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	row = statement.QueryRow(query, workflowID)
	workflow = &models.Workflow{}
	err = row.Scan(&workflow.Id, &workflow.Version, &workflow.Name, &workflow.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			// no workflow found
			return nil, nil
		}
		logger.ErrorF("Error scanning row: %v", err)
		err = errors.New("error scanning workflow")
		return
	}
	return
}

func (s *PostgresStorage) ListWorkflows() (workflows []*models.Workflow, err error) {
	query := `SELECT * FROM workflows`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	rows, err := statement.Query(query)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error fetching workflows")
		return
	}
	defer rows.Close()
	workflows = make([]*models.Workflow, 0)
	for rows.Next() {
		workflow := &models.Workflow{}
		err = rows.Scan(&workflow.Id, &workflow.Version, &workflow.Name, &workflow.Description)
		if err != nil {
			if err == sql.ErrNoRows {
				// no workflow found
				return nil, nil
			}
			logger.ErrorF("Error scanning row: %v", err)
			err = errors.New("error scanning workflow")
			return
		}
		workflows = append(workflows, workflow)
	}
	return
}

func (s *PostgresStorage) ListWorkflowVersions(workflowID string) (workflows []*models.Workflow, err error) {
	query := `SELECT * FROM workflows WHERE workflow_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	rows, err := statement.Query(query, workflowID)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error fetching workflow versions")
		return
	}
	defer rows.Close()
	workflows = make([]*models.Workflow, 0)
	for rows.Next() {
		workflow := &models.Workflow{}
		err = rows.Scan(&workflow.Id, &workflow.Version, &workflow.Name, &workflow.Description)
		if err != nil {
			if err == sql.ErrNoRows {
				// no workflow found
				return nil, nil
			}
			logger.ErrorF("Error scanning row: %v", err)
			err = errors.New("error scanning workflow")
			return
		}
	}
	return
}

func (s *PostgresStorage) ListActions() (actions []*models.ActionSpec, err error) {
	query := `SELECT * FROM actions`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	rows, err := statement.Query(query)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error fetching actions")
		return
	}
	defer rows.Close()
	actions = make([]*models.ActionSpec, 0)
	for rows.Next() {
		action := &models.ActionSpec{}
		err = rows.Scan(&action.Id, &action.Name, &action.Description, &action.Endpoint)
		if err != nil {
			if err == sql.ErrNoRows {
				// no action found
				return nil, nil
			}
			logger.ErrorF("Error scanning row: %v", err)
			err = errors.New("error scanning action")
			return
		}
		actions = append(actions, action)
	}
	return
}

func (s *PostgresStorage) SaveAction(action *models.ActionSpec) (err error) {
	query := `Insert into actions (id, name, description, endpoint) VALUES (?, ?, ?, ?)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	endpointJSON, err := json.Marshal(action.Endpoint)
	if err != nil {
		logger.ErrorF("Error marshalling endpoint: %v", err)
		err = errors.New("error marshalling endpoint")
		return
	}
	actionId := CreateId()
	_, err = statement.Exec(query, actionId, action.Name, action.Description, endpointJSON)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error saving action")
		return
	}
	return
}

func (s *PostgresStorage) LockInstance(instanceID string) (isLocked bool, err error) {
	query := `Update workflow_data set is_locked = true where instance_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, instanceID)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error locking instance")
		return
	}
	return
}

func (s *PostgresStorage) SaveStepChangeEvent(stepEvent *events.StepChangeEvent) (err error) {
	query := `INSERT INTO step_change_event (instance_id, event_id, status, data) VALUES (?, ?, ?, ?)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	stepEventDataJSON, err := json.Marshal(stepEvent.Data)
	if err != nil {
		logger.ErrorF("Error marshalling step event data: %v", err)
		err = errors.New("error marshalling step event data")
		return
	}
	_, err = statement.Exec(query, stepEvent.InstanceId, stepEvent.EventId, stepEvent.Status, stepEventDataJSON)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error saving step change event")
		return
	}
	return
}

func (s *PostgresStorage) SavePipeline(pipeline *data.Pipeline) (err error) {
	query := `UPDATE workflow_data SET pipeline_data = ? WHERE instance_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	mappedData := pipeline.Map()
	pipelineJSON, err := json.Marshal(mappedData)
	if err != nil {
		logger.ErrorF("Error marshalling pipeline data: %v", err)
		err = errors.New("error marshalling pipeline data")
		return
	}
	_, err = statement.Exec(query, pipelineJSON, pipeline.Id())
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error saving pipeline data")
		return
	}
	return
}

func (s *PostgresStorage) SaveState(workflowState *WorkflowState) (err error) {
	query := `Insert into workflow_state (instance_id, instance_version, workflow_id, workflow_version, status, error) VALUES (?, ?, ?, ?, ?, ?)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, workflowState.InstanceId, workflowState.InstanceVersion, workflowState.WorkflowId, workflowState.WorkflowVersion, workflowState.Status, workflowState.Error)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error saving workflow state")
		return
	}
	return
}

func (s *PostgresStorage) SaveStepState(stepState *StepState) (err error) {
	query := `INSERT INTO step_state (step_id, parent_step, child_count, status, instance_id) VALUES (?, ?, ?, ?, ?)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, stepState.StepId, stepState.ParentStep, stepState.ChildCount, stepState.Status, stepState.InstanceId)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error saving step state")
		return
	}
	return
}

func (s *PostgresStorage) SaveWorkflow(workflow *models.Workflow) (err error) {
	query := `INSERT INTO workflows (workflow_id, version, name, description) VALUES (?, ?, ?, ?)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, workflow.Id, workflow.Version, workflow.Name, workflow.Description)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error saving workflow")
		return
	}
	return
}

func (s *PostgresStorage) UnlockInstance(instanceID string) (err error) {
	query := `UPDATE workflow_data set is_locked = false where instance_id = ?`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement: %v", err)
		err = errors.New("error preparing statement")
		return
	}
	_, err = statement.Exec(query, instanceID)
	if err != nil {
		logger.ErrorF("Error executing query: %v", err)
		err = errors.New("error unlocking instance")
		return
	}
	return
}

func (s *PostgresStorage) PrepareStatement(query string) (*sql.Stmt, error) {
	return s.Database.Prepare(query)
}

func (s *PostgresStorage) Config() *config.StorageConfig {
	return &config.StorageConfig{
		Type: config.PostgresStorageType,
		Provider: &config.Provider{
			PostgreSQL: &config.PostgresStorage{},
		},
	}
}
