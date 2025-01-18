package runtime

import (
	"oss.nandlabs.io/orcaloop-sdk/models"
	"oss.nandlabs.io/orcaloop/storage"

	"oss.nandlabs.io/orcaloop-sdk/utils"
)

type WorkflowManager struct {
	store storage.Storage
}

func (wfm *WorkflowManager) Register(workflow *models.Workflow) (err error) {

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
