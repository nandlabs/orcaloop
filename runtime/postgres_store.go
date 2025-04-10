package runtime

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"oss.nandlabs.io/golly/codec"
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
		if c.Provider.PostgreSQL.Schema == "" {
			c.Provider.PostgreSQL.Schema = "public"
		}
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s", c.Provider.PostgreSQL.Host, c.Provider.PostgreSQL.Port, c.Provider.PostgreSQL.User, c.Provider.PostgreSQL.Password, c.Provider.PostgreSQL.Database, c.Provider.PostgreSQL.SSLMode, c.Provider.PostgreSQL.Schema)
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
	query := `SELECT id, name, description, endpoint FROM actions WHERE id = $1 AND is_deleted = $2`
	sqlStatement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to fetch action spec: %v", err)
		err = errors.New("error preparing statement to fetch action spec")
		return
	}
	row := sqlStatement.QueryRow(id, false)
	actionSpec = &models.ActionSpec{}
	var endpointJSON []byte
	err = row.Scan(&actionSpec.Id, &actionSpec.Name, &actionSpec.Description, &endpointJSON)
	if err != nil {
		logger.Error("Error scanning action spec: %v", err)
		err = errors.New("error scanning action spec")
		return
	}
	err = codec.JsonCodec().DecodeBytes(endpointJSON, &actionSpec.Endpoint)
	if err != nil {
		logger.ErrorF("Error unmarshalling action spec endpoint: %v", err)
		err = errors.New("error unmarshalling action spec endpoint")
		return
	}
	return
}

func (s *PostgresStorage) ActionSpecs() (actionSpecs []*models.ActionSpec, err error) {
	query := `SELECT * FROM actions WHERE is_deleted = $1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to fetch action specs: %v", err)
		err = errors.New("error preparing statement to fetch action specs")
		return
	}
	rows, err := statement.Query(false)
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
			logger.ErrorF("Error scanning row to fetch action specs: %v", err)
			err = errors.New("error scanning action spec row")
			return
		}
		actionSpecs = append(actionSpecs, actionSpec)
	}
	return
}

func (s *PostgresStorage) AddPendingSteps(instanceId string, pendingStep ...*PendingStep) (err error) {
	query := `INSERT INTO pending_steps (id, instance_id, step_id, data) VALUES ($1, $2, $3, $4)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to add pending step: %v", err)
		err = errors.New("error preparing statement to add pending step")
		return
	}
	for _, step := range pendingStep {
		var pendingStepJSON []byte
		pendingStepJSON, err = codec.JsonCodec().EncodeToBytes(step)
		if err != nil {
			logger.ErrorF("Error marshalling pending step: %v", err)
			err = errors.New("error marshalling pending step")
			return
		}
		_, err = statement.Exec(step.Id, instanceId, step.StepId, pendingStepJSON)
		if err != nil {
			logger.ErrorF("Error executing query: %v", err)
			err = errors.New("error adding pending step/s")
			return
		}
	}
	return
}

func (s *PostgresStorage) ActionEndpoint(id string) (endpoint *models.Endpoint, err error) {
	query := `SELECT endpoint FROM actions WHERE id = $1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to fetch action endpoint: %v", err)
		err = errors.New("error preparing statement to fetch action endpoint")
		return
	}
	row := statement.QueryRow(id)
	var endpointJSON []byte
	err = row.Scan(&endpointJSON)
	if err != nil {
		logger.ErrorF("Error scanning row for action endpoint: %v", err)
		err = errors.New("error scanning action endpoint")
		return
	}
	err = codec.JsonCodec().DecodeBytes(endpointJSON, &endpoint)
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
	query := `INSERT INTO workflow_data (instance_id, workflow_id, workflow_version, pipeline_data) VALUES ($1, $2, $3, $4)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to create a new instance: %v", err)
		err = errors.New("error preparing statement to create a new instance")
		return
	}
	mappedData := pipeline.Map()
	pipelineJSON, err := codec.JsonCodec().EncodeToBytes(mappedData)
	if err != nil {
		logger.ErrorF("Error marshalling pipeline data: %v", err)
		err = errors.New("error marshalling pipeline data")
		return
	}
	logger.InfoF("Creating new instance with ID: %s, Workflow ID: %s, Version: %d", instanceID, workflowID, pipeline.GetWorkflowVersion())
	_, err = statement.Exec(instanceID, workflowID, pipeline.GetWorkflowVersion(), pipelineJSON)
	if err != nil {
		logger.ErrorF("Error executing query to create new instance: %v", err)
		err = errors.New("error creating new instance")
		return
	}
	return
}

func (s *PostgresStorage) DeleteAction(id string) (err error) {
	query := `UPDATE actions set is_deleted = $1 WHERE id = $2`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to delete action: %v", err)
		err = errors.New("error preparing statement to delete action")
		return
	}
	_, err = statement.Exec(true, id)
	if err != nil {
		logger.ErrorF("Error executing query to delete action: %v", err)
		err = errors.New("error deleting action")
		return
	}
	return
}

func (s *PostgresStorage) DeletePendingStep(instanceID string, pendingStep *PendingStep) (err error) {
	query := `Delete from pending_steps WHERE id = $1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to delete pending step: %v", err)
		err = errors.New("error preparing statement to delete pending step")
		return
	}
	_, err = statement.Exec(pendingStep.Id)
	if err != nil {
		logger.ErrorF("Error executing query to delete pending step: %v", err)
		err = errors.New("error deleting pending step")
		return
	}
	return nil
}

func (s *PostgresStorage) DeleteWorkflow(workflowID string, version int) (err error) {
	query := `Update workflows set is_deleted = $1 WHERE workflow_id = $2 AND version = $3`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to delete a workflow: %v", err)
		err = errors.New("error preparing statement to delete a workflow")
		return
	}
	_, err = statement.Exec(true, workflowID, version)
	if err != nil {
		logger.ErrorF("Error executing query to delete a workflow: %v", err)
		err = errors.New("error deleting workflow")
		return
	}
	return
}

func (s *PostgresStorage) DeleteStepChangeEvent(instanceID, eventID string) (err error) {
	// soft delete
	query := `UPDATE step_change_event SET is_deleted = $1 WHERE instance_id = $2 AND event_id = $3`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to delete step change event: %v", err)
		err = errors.New("error preparing statement to delete step change event")
		return
	}
	_, err = statement.Exec(true, instanceID, eventID)
	if err != nil {
		logger.ErrorF("Error executing query to delete step change event: %v", err)
		err = errors.New("error deleting step change event")
		return
	}
	return nil
}

func (s *PostgresStorage) GetPipeline(id string) (pipelineData *data.Pipeline, err error) {
	query := `SELECT pipeline_data FROM workflow_data WHERE instance_id = $1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to fetch pipeline data: %v", err)
		err = errors.New("error preparing statement to fetch pipeline data")
		return
	}
	row := statement.QueryRow(id)
	var pipelineJSON []byte
	err = row.Scan(&pipelineJSON)
	if err != nil {
		logger.ErrorF("Error scanning row for pipeline data: %v", err)
		err = errors.New("error scanning pipeline data")
		return
	}
	var pipelineDataMap map[string]any
	err = codec.JsonCodec().DecodeBytes(pipelineJSON, &pipelineDataMap)
	if err != nil {
		logger.ErrorF("Error unmarshalling pipeline data: %v", err)
		err = errors.New("error unmarshalling pipeline data")
		return
	}
	pipelineData = data.NewPipelineFrom(pipelineDataMap)
	return
}

func (s *PostgresStorage) GetState(instanceID string) (state *WorkflowState, err error) {
	query := `SELECT instance_id, workflow_id, workflow_version, instance_version, status FROM workflow_state WHERE instance_id = $1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to get workflow state: %v", err)
		err = errors.New("error preparing statement to get workflow state")
		return
	}
	row := statement.QueryRow(instanceID)
	state = &WorkflowState{}
	var status string
	err = row.Scan(&state.InstanceId, &state.WorkflowId, &state.WorkflowVersion, &state.InstanceVersion, &status)
	if err != nil {
		logger.ErrorF("Error scanning row for workflow state: %v", err)
		err = errors.New("error scanning workflow state")
		return
	}
	state.Status = models.StringToStatus[status]
	return
}

func (s *PostgresStorage) GetAndRemoveNextPendingStep(instanceID string) (pendingStep *PendingStep, err error) {
	// fetch the earliest pending step based on the created timestamp
	query := `SELECT data FROM pending_steps WHERE instance_id = $1 AND is_deleted = $2 ORDER BY created_at ASC LIMIT 1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to get next pending step: %v", err)
		err = errors.New("error preparing statement to get next pending step")
		return
	}
	row := statement.QueryRow(instanceID, false)
	pendingStep = &PendingStep{}
	var pendingStepJSON []byte
	err = row.Scan(&pendingStepJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// no pending steps found
			return nil, nil
		}
		logger.ErrorF("Error scanning row for pending step: %v", err)
		err = errors.New("error scanning pending step")
		return
	}
	err = codec.JsonCodec().DecodeBytes(pendingStepJSON, pendingStep)
	if err != nil {
		logger.ErrorF("Error unmarshalling pending step: %v", err)
		err = errors.New("error unmarshalling pending step")
		return
	}
	query = `UPDATE pending_steps SET is_deleted = $1 WHERE id = $2`
	statement, err = s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to delete pending step: %v", err)
		err = errors.New("error preparing statement to delete pending step")
		return
	}
	_, err = statement.Exec(true, pendingStep.Id)
	if err != nil {
		logger.ErrorF("Error executing query to delete pending step: %v", err)
		err = errors.New("error deleting pending step")
		return
	}
	return
}

func (s *PostgresStorage) GetPendingSteps(instanceID string) (pendingSteps []*PendingStep, err error) {
	query := `SELECT data FROM pending_steps WHERE instance_id = $1 AND status != 'Completed' AND is_deleted = $2`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to get pending steps: %v", err)
		err = errors.New("error preparing statement to get pending steps")
		return
	}
	rows, err := statement.Query(instanceID, false)
	if err != nil {
		logger.ErrorF("Error executing query to fetch pending steps: %v", err)
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
			logger.ErrorF("Error scanning row for pending step: %v", err)
			err = errors.New("error scanning pending step")
			return
		}
		err = codec.JsonCodec().DecodeBytes(pendingStepJSON, pendingStep)
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
	query := `SELECT instance_id, event_id, status, data  FROM step_change_event WHERE instance_id = $1 AND is_deleted = $2`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to get step change events: %v", err)
		err = errors.New("error preparing statement to get step change events")
		return
	}
	rows, err := statement.Query(instanceID, false)
	if err != nil {
		logger.ErrorF("Error executing query to get step change events: %v", err)
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
			logger.ErrorF("Error scanning row for step change event: %v", err)
			err = errors.New("error scanning step change event")
			return
		}
		err = codec.JsonCodec().DecodeBytes(stepChangeEventDataJSON, &stepChangeEvent.Data)
		if err != nil {
			logger.ErrorF("Error unmarshalling step change event data: %v", err)
			err = errors.New("error unmarshalling step change event data")
			return
		}
		stepChangeEvents = append(stepChangeEvents, stepChangeEvent)
	}
	return
}

func (s *PostgresStorage) GetStepState(instanceID, stepID string, iteration int) (stepState *StepState, err error) {
	query := `SELECT step_id, parent_step, child_count, status, instance_id, iteration FROM step_state WHERE instance_id = $1 AND step_id = $2 AND iteration = $3`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to get step state: %v", err)
		err = errors.New("error preparing statement to get step state")
		return
	}
	row := statement.QueryRow(instanceID, stepID, iteration)
	stepState = &StepState{}
	var status string
	err = row.Scan(&stepState.StepId, &stepState.ParentStep, &stepState.ChildCount, &status, &stepState.InstanceId, &stepState.Iteration)
	if err != nil {
		if err == sql.ErrNoRows {
			// no step state found
			return nil, nil
		}
		logger.ErrorF("Error scanning row for step state: %v", err)
		err = errors.New("error scanning step state")
		return
	}
	stepState.Status = models.StringToStatus[status]
	return
}

func (s *PostgresStorage) GetStepStates(instanceID string) (stepState map[string][]*StepState, err error) {
	query := `SELECT step_state FROM step_state WHERE instance_id = $1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to get step states: %v", err)
		err = errors.New("error preparing statement to get step states")
		return
	}
	rows, err := statement.Query(instanceID)
	if err != nil {
		logger.ErrorF("Error executing query to fetch step states: %v", err)
		err = errors.New("error fetching step states")
		return
	}
	defer rows.Close()
	stepState = make(map[string][]*StepState)
	for rows.Next() {
		stepStateItem := &StepState{}
		var stepStateJSON []byte
		err = rows.Scan(&stepStateJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				// no step state found
				return nil, nil
			}
			logger.ErrorF("Error scanning row for step state: %v", err)
			err = errors.New("error scanning step state")
			return
		}
		err = codec.JsonCodec().DecodeBytes(stepStateJSON, stepStateItem)
		if err != nil {
			logger.ErrorF("Error unmarshalling step state: %v", err)
			err = errors.New("error unmarshalling step state")
			return
		}
		if _, ok := stepState[stepStateItem.StepId]; !ok {
			stepState[stepStateItem.StepId] = make([]*StepState, 0)
		}
		stepState[stepStateItem.StepId] = append(stepState[stepStateItem.StepId], stepStateItem)
	}
	return
}

func (s *PostgresStorage) GetWorkflow(workflowID string, version int) (workflow *models.Workflow, err error) {
	query := `SELECT workflow_document FROM workflows WHERE workflow_id = $1 AND version = $2`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to fetch workflow: %v", err)
		err = errors.New("error preparing statement to fetch workflow")
		return
	}
	row := statement.QueryRow(workflowID, version)
	workflow = &models.Workflow{}
	var workflowDocumentJSON []byte
	err = row.Scan(&workflowDocumentJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// no workflow found
			err = errors.New("workflow not found")
			return
		}
		logger.ErrorF("Error scanning row for workflow: %v", err)
		err = errors.New("error scanning workflow")
		return
	}
	err = codec.JsonCodec().DecodeBytes(workflowDocumentJSON, workflow)
	if err != nil {
		logger.ErrorF("Error unmarshalling workflow document: %v", err)
		err = errors.New("error unmarshalling workflow document")
		return
	}
	return
}

func (s *PostgresStorage) GetWorkflowByInstance(instanceID string) (workflow *models.Workflow, err error) {
	query := `select workflow_id, workflow_version from workflow_data where instance_id = $1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement fetch workflow by instance: %v", err)
		err = errors.New("error preparing statement to fetch workflow by instance")
		return
	}
	row := statement.QueryRow(instanceID)
	var workflowID, workflowVersion string
	err = row.Scan(&workflowID, &workflowVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			// no workflow found
			return nil, nil
		}
		logger.ErrorF("Error scanning row workflowID: %v", err)
		err = errors.New("error scanning workflowID")
		return
	}
	query = `select workflow_document from workflows where workflow_id = $1 AND version = $2`
	statement, err = s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to fetch workflows by workflowID: %v", err)
		err = errors.New("error preparing statement to fetch workflows by workflowID")
		return
	}
	row = statement.QueryRow(workflowID, workflowVersion)
	workflow = &models.Workflow{}
	var workflowDocumentJSON []byte
	err = row.Scan(&workflowDocumentJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// no workflow found
			return nil, nil
		}
		logger.ErrorF("Error scanning row for workflow: %v", err)
		err = errors.New("error scanning workflow")
		return
	}
	err = codec.JsonCodec().DecodeBytes(workflowDocumentJSON, workflow)
	if err != nil {
		logger.ErrorF("Error unmarshalling workflow document: %v", err)
		err = errors.New("error unmarshalling workflow document")
		return
	}
	return
}

func (s *PostgresStorage) ListWorkflows() (workflows []*models.Workflow, err error) {
	query := `SELECT * FROM workflows`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to fetch workflows: %v", err)
		err = errors.New("error preparing statement to fetch workflows")
		return
	}
	rows, err := statement.Query()
	if err != nil {
		logger.ErrorF("Error executing query to list workflows: %v", err)
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
			logger.ErrorF("Error scanning row for workflow: %v", err)
			err = errors.New("error scanning workflow")
			return
		}
		workflows = append(workflows, workflow)
	}
	return
}

func (s *PostgresStorage) ListWorkflowVersions(workflowID string) (workflows []*models.Workflow, err error) {
	query := `SELECT * FROM workflows WHERE workflow_id = $1`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to list workflow version: %v", err)
		err = errors.New("error preparing statement to list workflow version")
		return
	}
	rows, err := statement.Query(workflowID)
	if err != nil {
		logger.ErrorF("Error executing query workflow versions: %v", err)
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
			logger.ErrorF("Error scanning row workflow versions: %v", err)
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
		logger.ErrorF("Error preparing statement to list actions: %v", err)
		err = errors.New("error preparing statement to list actions")
		return
	}
	rows, err := statement.Query()
	if err != nil {
		logger.ErrorF("Error executing query to list actions: %v", err)
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
			logger.ErrorF("Error scanning row for action: %v", err)
			err = errors.New("error scanning action")
			return
		}
		actions = append(actions, action)
	}
	return
}

func (s *PostgresStorage) SaveAction(action *models.ActionSpec) (err error) {
	checkQuery := `SELECT COUNT(*) FROM actions WHERE id = $1`
	statement, err := s.PrepareStatement(checkQuery)
	if err != nil {
		logger.ErrorF("Error preparing statement to check action existence: %v", err)
		err = errors.New("error preparing statement to check action existence")
		return
	}
	var count int
	err = statement.QueryRow(action.Id).Scan(&count)
	if err != nil {
		logger.ErrorF("Error executing query to check action existence: %v", err)
		err = errors.New("error checking action existence")
		return
	}
	if count > 0 {
		logger.InfoF("Action with ID %s already exists", action.Id)
		return
	}
	query := `Insert into actions (id, name, description, endpoint) VALUES ($1, $2, $3, $4)`
	statement, err = s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to save action: %v", err)
		err = errors.New("error preparing statement to save action")
		return
	}
	endpointJSON, err := codec.JsonCodec().EncodeToBytes(action.Endpoint)
	if err != nil {
		logger.ErrorF("Error marshalling endpoint: %v", err)
		err = errors.New("error marshalling endpoint")
		return
	}
	_, err = statement.Exec(action.Id, action.Name, action.Description, endpointJSON)
	if err != nil {
		logger.ErrorF("Error executing query to save action: %v", err)
		err = errors.New("error saving action")
		return
	}
	return
}

func (s *PostgresStorage) LockInstance(instanceID string) (isLocked bool, err error) {
	query := `Update workflow_data set is_locked = $1 where instance_id = $2`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to lock instance: %v", err)
		err = errors.New("error preparing statement to lock instance")
		return
	}
	_, err = statement.Exec(true, instanceID)
	if err != nil {
		logger.ErrorF("Error executing query to lock instance: %v", err)
		err = errors.New("error locking instance")
		return
	}
	isLocked = true
	return
}

func (s *PostgresStorage) SaveStepChangeEvent(stepEvent *events.StepChangeEvent) (err error) {
	query := `INSERT INTO step_change_event (instance_id, event_id, step_id, status, data) VALUES ($1, $2, $3, $4, $5)`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to save step change event: %v", err)
		err = errors.New("error preparing statement to save step change event")
		return
	}
	stepEventDataJSON, err := codec.JsonCodec().EncodeToBytes(stepEvent.Data)
	if err != nil {
		logger.ErrorF("Error marshalling step event data: %v", err)
		err = errors.New("error marshalling step event data")
		return
	}
	_, err = statement.Exec(stepEvent.InstanceId, stepEvent.EventId, stepEvent.StepId, stepEvent.Status.String(), stepEventDataJSON)
	if err != nil {
		logger.ErrorF("Error executing query to save step change event: %v", err)
		err = errors.New("error saving step change event")
		return
	}
	return
}

func (s *PostgresStorage) SavePipeline(pipeline *data.Pipeline) (err error) {
	// query := `Insert into workflow_data `
	query := `UPDATE workflow_data SET pipeline_data = $1 WHERE instance_id = $2`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to save pipeline: %v", err)
		err = errors.New("error preparing statement to save pipeline")
		return
	}
	mappedData := pipeline.Map()
	pipelineJSON, err := codec.JsonCodec().EncodeToBytes(mappedData)
	if err != nil {
		logger.ErrorF("Error marshalling pipeline data: %v", err)
		err = errors.New("error marshalling pipeline data")
		return
	}
	_, err = statement.Exec(pipelineJSON, pipeline.Id())
	if err != nil {
		logger.ErrorF("Error executing query to save pipeline: %v", err)
		err = errors.New("error saving pipeline data")
		return
	}
	return
}

func (s *PostgresStorage) SaveState(workflowState *WorkflowState) (err error) {
	query := `Insert into workflow_state (instance_id, workflow_id, workflow_version, instance_version, status) VALUES ($1, $2, $3, $4, $5) on conflict(instance_id, workflow_id, workflow_version) do update set instance_version=$4, status=$5`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to save state: %v", err)
		err = errors.New("error preparing statement to save state")
		return
	}
	logger.InfoF("Saving state: %v", workflowState)
	_, err = statement.Exec(workflowState.InstanceId, workflowState.WorkflowId, workflowState.WorkflowVersion, workflowState.InstanceVersion, workflowState.Status.String())
	if err != nil {
		logger.ErrorF("Error executing query to save state: %v", err)
		err = errors.New("error saving workflow state")
		return
	}
	return
}

func (s *PostgresStorage) SaveStepState(stepState *StepState) (err error) {
	query := `INSERT INTO step_state (instance_id, step_id, iteration, parent_step, child_count, status, step_state) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		ON Conflict(instance_id, step_id, iteration) 
		DO UPDATE set parent_step=$4, child_count=$5, status=$6, step_state=$7`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to save step state: %v", err)
		err = errors.New("error preparing statement to save step state")
		return
	}
	var stepStateJSON []byte
	stepStateJSON, err = codec.JsonCodec().EncodeToBytes(stepState)
	if err != nil {
		logger.ErrorF("Error marshalling step state data: %v", err)
		err = errors.New("error marshalling step state data")
		return
	}
	_, err = statement.Exec(stepState.InstanceId, stepState.StepId, stepState.Iteration, stepState.ParentStep, stepState.ChildCount, stepState.Status.String(), stepStateJSON)
	if err != nil {
		logger.ErrorF("Error executing query to save step state: %v", err)
		err = errors.New("error saving step state")
		return
	}
	return
}

func (s *PostgresStorage) SaveWorkflow(workflow *models.Workflow) (err error) {
	checkQuery := `SELECT COUNT(*) FROM workflows WHERE workflow_id = $1 AND version = $2`
	statement, err := s.PrepareStatement(checkQuery)
	if err != nil {
		logger.ErrorF("Error preparing statement to check workflow existence: %v", err)
		err = errors.New("error preparing statement to check workflow existence")
		return
	}
	var count int
	err = statement.QueryRow(workflow.Id, workflow.Version).Scan(&count)
	if err != nil {
		logger.ErrorF("Error executing query to check workflow existence: %v", err)
		err = errors.New("error checking workflow existence")
		return
	}
	if count > 0 {
		logger.InfoF("Workflow with ID %s and version %d already exists", workflow.Id, workflow.Version)
		return
	}
	query := `INSERT INTO workflows (workflow_id, version, name, description, workflow_document) VALUES ($1, $2, $3, $4, $5)`
	statement, err = s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to save workflow: %v", err)
		err = errors.New("error preparing statement to save workflow")
		return
	}
	workflowDocumentJSON, err := codec.JsonCodec().EncodeToBytes(workflow)
	if err != nil {
		logger.ErrorF("Error marshalling workflow document: %v", err)
		err = errors.New("error marshalling workflow document")
		return
	}
	_, err = statement.Exec(workflow.Id, workflow.Version, workflow.Name, workflow.Description, workflowDocumentJSON)
	if err != nil {
		logger.ErrorF("Error executing query to save workflow: %v", err)
		err = errors.New("error saving workflow")
		return
	}
	return
}

func (s *PostgresStorage) UnlockInstance(instanceID string) (err error) {
	query := `UPDATE workflow_data set is_locked = $1 where instance_id = $2`
	statement, err := s.PrepareStatement(query)
	if err != nil {
		logger.ErrorF("Error preparing statement to unlock instance: %v", err)
		err = errors.New("error preparing statement to unlock instance")
		return
	}
	_, err = statement.Exec(false, instanceID)
	if err != nil {
		logger.ErrorF("Error executing query to unlock instance: %v", err)
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
