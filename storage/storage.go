package storage

import (
	"fmt"

	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop/actions"
	"oss.nandlabs.io/orcaloop/models"
)

// ErrWfNotFound is an error that is returned when the requested item is not found
// This error is returned when the requested item is not found
var ErrWorkFlowNotFound = func(id string) error { return fmt.Errorf("workflow with id %s not found ", id) }

type Storage interface {
	// ActionSpec returns the spec of the action
	ActionSpec(id string) (*actions.ActionSpec, error)
	// ActionSpecs returns a list of action specs
	ActionSpecs() ([]*actions.ActionSpec, error)
	// Archive archives a workflow configuration
	ArchiveInstance(workflowID string, archiveInstance bool) error
	// CreateNewInstance creates a new instance
	CreateNewInstance(workflowID string, instanceID string, pipeline *data.Pipeline) error
	// DeleteAction deletes the action
	DeleteAction(id string) error
	// GetPipeline retrieves the pipeline configuration of a workflow
	GetPipeline(id string) (*data.Pipeline, error)
	//GetState retrieves the state of a workflow
	GetState(instanceId string) (*models.WorkflowState, error)
	//GetStepChangeEvent retrieves the state change events
	GetStepChangeEvents(instanceId string) ([]*models.StepChangeEvent, error)
	//GetStepContext provides step context
	GetStepState(instanceId, stepId string) (*models.StepState, error)
	// GetWorkflow retrieves a stored workflow configuration
	GetWorkflow(workflowID string, version int) (*models.Workflow, error)
	// GetWorkflowByInstance Id retrieves a stored workflow configuration
	GetWorkflowByInstance(id string) (*models.Workflow, error)
	// ListActions returns a list of all actions
	ListActions() ([]*actions.ActionSpec, error)
	// LockInstance locks an instance
	LockInstance(id string) (bool, error)
	// SaveAction saves the action
	SaveAction(action *actions.ActionSpec) error
	// SaveStepChangeEvent saves the step change event
	SaveStepChangeEvent(stepEvent *models.StepChangeEvent) error
	// SavePipeline updates the pipeline configuration of a workflow
	SavePipeline(pipeline *data.Pipeline) error
	// SaveState updates the state of a workflow
	SaveState(workflowState *models.WorkflowState) error
	// SaveStepState Saves the step state. If the step state does not exist, it creates a new one
	SaveStepState(stepState *models.StepState) error
	// SaveWorkflow stores the workflow configuration
	SaveWorkflow(workflow *models.Workflow) error
	// UnlockInstance unlocks an instance
	UnlockInstance(id string) error
}
