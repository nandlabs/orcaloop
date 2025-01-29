package api

import (
	"oss.nandlabs.io/orcaloop-sdk/models"
)

type StartWorkflowRequest struct {
	// WorkflowId is the id of the workflow to start
	WorkflowId string `json:"workflowId" yaml:"workflowId"`
	// Version is the version of the workflow to start
	Version int `json:"version" yaml:"version"`
	// Input is the input to the workflow
	Input map[string]any `json:"input" yaml:"input"`
}

type StartWorkflowResponse struct {
	*APIBaseResponse
	// InstanceId is the id of the workflow instance
	InstanceId string `json:"instanceId,omitempty" yaml:"instanceId,omitempty"`
}

type GetWorkflowResponse struct {
	*APIBaseResponse
	// Workflow is the workflow
	Workflow *models.Workflow `json:"workflow,omitempty" yaml:"workflow,omitempty"`
}

type GetWorkflowsResponse struct {
	*APIBaseResponse
	// Workflows is the list of workflows
	Workflows []*models.Workflow `json:"workflows,omitempty" yaml:"workflows,omitempty"`
}

type WorkflowStatusReqeust struct {
	// InstanceId is the id of the workflow instance
	InstanceId string `json:"instanceId" yaml:"instanceId"`
}

type WorkflowStatusResponse struct {
	*APIBaseResponse
	// Status is the status of the workflow instance
	Status string `json:"status,omitempty" yaml:"status,omitempty"`
	//pipeline is the data of the workflow instance
	Pipeline map[string]any `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
}

// GetActionsResponse is the response for GetActions
type GetActionsResponse struct {
	*APIBaseResponse
	// Actions is the list of actions
	ActionSpec []*models.ActionSpec `json:"action_specs,omitempty" yaml:"action_specs,omitempty"`
}

// GetActionResponse is the response for GetAction
type GetActionResponse struct {
	*APIBaseResponse
	// Action is the action
	ActionSpec *models.ActionSpec `json:"action_spec,omitempty" yaml:"action_spec,omitempty"`
}
