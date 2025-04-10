package runtime

import (
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"

	"oss.nandlabs.io/orcaloop-sdk/utils"
)

type WorkflowManager struct {
	store Storage
}

// NewWorkflowManager creates a new WorkflowManager with the given storage.
func NewWorkflowManager(store Storage) *WorkflowManager {
	return &WorkflowManager{store: store}
}

// GetWorkflow retrieves a workflow from the WorkflowManager.
// It takes the ID and version of the workflow to retrieve.
// It returns the workflow if it exists, otherwise it returns an error.
func (wfm *WorkflowManager) GetWorkflow(id string, version int) (workflow *models.Workflow, err error) {

	workflow, err = wfm.store.GetWorkflow(id, version)
	if err != nil {
		return
	}
	return
}

// GetWorkflows returns a list of all workflows registered in the WorkflowManager.
// It returns an error if the workflows could not be retrieved.
func (wfm *WorkflowManager) GetWorkflows() (workflows []*models.Workflow, err error) {

	workflows, err = wfm.store.ListWorkflows()

	return
}

// DeleteWorkflow removes a workflow from the WorkflowManager.
// It takes the ID and version of the workflow to delete.
// It returns an error if the workflow could not be deleted.
func (wfm *WorkflowManager) DeleteWorkflow(id string, version int) (err error) {

	err = wfm.store.DeleteWorkflow(id, version)

	return
}

// Save registers a new workflow in the WorkflowManager.
// It first checks if the workflow is already registered by querying the store with the workflow's ID and version.
// If the workflow is already registered, it returns an ErrWorkflowAlreadyRegistered error.
// If the workflow is not registered, it validates the workflow using utils.ValidateWorkflow.
// If the validation passes, it saves the workflow to the store.
// It returns an error if any of these steps fail.
func (wfm *WorkflowManager) Save(workflow *models.Workflow) (err error) {

	// Check if workflow is already registered
	_, err = wfm.store.GetWorkflow(workflow.Id, workflow.Version)
	if err == nil {
		err = ErrWorkflowAlreadyRegistered(workflow.Id, workflow.Version)
		return
	}
	// Validate workflow
	err = utils.ValidateWorkflow(*workflow)
	if err != nil {
		return
	}

	// Save workflow
	err = wfm.store.SaveWorkflow(workflow)

	return
}

// Start initializes and starts the execution of a workflow with the given ID and version.
// It takes an input map containing the initial data for the workflow.
//
// Parameters:
//   - id: The unique identifier of the workflow to start.
//   - version: The version of the workflow to start.
//   - input: A map containing the initial data for the workflow.
//
// Returns:
//   - err: An error if the workflow could not be started, otherwise nil.
func (wfm *WorkflowManager) Start(id string, version int, input map[string]any) (instanceId string, err error) {

	// Get workflow
	workflow, err := wfm.GetWorkflow(id, version)
	if err != nil {
		return
	}

	instanceId = CreateId()
	// Create pipeline
	pipeline := data.NewPipelineFrom(input)
	pipeline.Set(data.InstanceIdKey, instanceId)
	pipeline.Set(data.WorkflowIdKey, id)
	pipeline.Set(data.WorkflowVersionKey, version)
	// Save pipeline
	// err = wfm.store.SavePipeline(pipeline)
	err = wfm.store.CreateNewInstance(id, instanceId, pipeline)
	if err != nil {
		return
	}
	workflowState := &WorkflowState{
		InstanceId:      instanceId,
		InstanceVersion: 1,
		WorkflowId:      id,
		WorkflowVersion: version,
		Status:          models.StatusRunning,
	}
	// Save workflow state
	err = wfm.store.SaveState(workflowState)
	if err != nil {
		return
	}
	// Create Workflow Executor
	wfe := WorkflowExecutor{storage: wfm.store}
	// Execute workflow
	err = wfe.Execute(workflow, pipeline)
	return
}
