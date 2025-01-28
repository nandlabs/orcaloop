package runtime

import (
	"oss.nandlabs.io/golly/managers"
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"
	"oss.nandlabs.io/orcaloop/config"

	"oss.nandlabs.io/orcaloop-sdk/events"
)

var StorageManager managers.ItemManager[Storage] = managers.NewItemManager[Storage]()

type Storage interface {
	Config() *config.StorageConfig
	ActionEndpoint(id string) (*models.Endpoint, error)
	// ActionSpec returns the spec of the action
	ActionSpec(id string) (*models.ActionSpec, error)
	// ActionSpecs returns a list of action specs
	ActionSpecs() ([]*models.ActionSpec, error)
	// Archive archives a workflow configuration
	ArchiveInstance(workflowID string, archiveInstance bool) error
	// CreateNewInstance creates a new instance
	CreateNewInstance(workflowID string, instanceID string, pipeline *data.Pipeline) error
	// DeleteAction deletes the action
	DeleteAction(id string) error
	// Delete Workflow deletes a workflow configuration
	DeleteWorkflow(workflowID string, version int) error
	// DeleteStepChangeEvent deletes the step change event
	DeleteStepChangeEvent(instanceId, eventId string) error
	// GetPipeline retrieves the pipeline configuration of a workflow
	GetPipeline(id string) (*data.Pipeline, error)
	//GetState retrieves the state of a workflow
	GetState(instanceId string) (*WorkflowState, error)
	//GetStepChangeEvent retrieves the state change events
	GetStepChangeEvents(instanceId string) ([]*events.StepChangeEvent, error)
	//GetStepContext provides step context
	GetStepState(instanceId, stepId string) (*StepState, error)
	// Get StepStates retrieves the states of all steps in a workflow
	GetStepStates(instanceId string) (map[string]*StepState, error)
	// GetWorkflow retrieves a stored workflow configuration
	GetWorkflow(workflowID string, version int) (*models.Workflow, error)
	// GetWorkflowByInstance Id retrieves a stored workflow configuration
	GetWorkflowByInstance(id string) (*models.Workflow, error)
	// ListWorkflows returns a list of all workflows
	ListWorkflows() ([]*models.Workflow, error)
	// ListActions returns a list of all actions
	ListActions() ([]*models.ActionSpec, error)
	// LockInstance locks an instance
	LockInstance(id string) (bool, error)
	// SaveAction saves the action
	SaveAction(action *models.ActionSpec) error
	// SaveStepChangeEvent saves the step change event
	SaveStepChangeEvent(stepEvent *events.StepChangeEvent) error
	// SavePipeline updates the pipeline configuration of a workflow
	SavePipeline(pipeline *data.Pipeline) error
	// SaveState updates the state of a workflow
	SaveState(workflowState *WorkflowState) error
	// SaveStepState Saves the step state. If the step state does not exist, it creates a new one
	SaveStepState(stepState *StepState) error
	// SaveWorkflow stores the workflow configuration
	SaveWorkflow(workflow *models.Workflow) error
	// UnlockInstance unlocks an instance
	UnlockInstance(id string) error
}
