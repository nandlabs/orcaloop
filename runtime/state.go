package runtime

import (
	"oss.nandlabs.io/orcaloop-sdk/data"
	"oss.nandlabs.io/orcaloop-sdk/models"
)

// WorkflowState represents the state of a workflow at a given point in time.
// It includes information such as the workflow's ID, version, status, the workflow itself,
// the states of individual steps within the workflow, and any error that may have occurred.
//
// Fields:
// - InstanceId: The unique identifier of the instance.
// - InstanceVersion: The version of the instance.
// - WorkflowId: The unique identifier of the workflow.
// - WorkflowVersion: The version of the workflow.
// - Status: The current status of the workflow.
// - Error: Any error that may have occurred during the execution of the workflow.

type WorkflowState struct {
	InstanceId      string        `json:"id" yaml:"id"`
	InstanceVersion int           `json:"version" yaml:"version"`
	WorkflowId      string        `json:"workflow_id" yaml:"workflow_id"`
	WorkflowVersion int           `json:"workflow_version" yaml:"workflow_version"`
	Status          models.Status `json:"status" yaml:"status"`
	Error           string        `json:"error" yaml:"error"`
}

// StepState represents the state of a step in a pipeline execution.
// It includes information about the instance, step identifiers, parent-child relationships,
// status, and input/output data of the step.
//
// Fields:
// - InstanceId: The unique identifier of the instance.
// - StepId: The unique identifier of the step.
// - ParentStep: The identifier of the parent step, if any.
// - ChildCount: The number of child steps associated with this step.
// - Status: The current status of the step.
// - Input: The input data for the step, represented as a Pipeline object.
// - Output: The output data from the step, represented as a Pipeline object.
type StepState struct {
	InstanceId string         `json:"instance_id" yaml:"instance_id"`
	StepId     string         `json:"step_id" yaml:"step_id"`
	Iteration  int            `json:"iteration" yaml:"iteration"`
	ParentStep string         `json:"parent_step" yaml:"parent_step"`
	ChildCount int            `json:"child_count" yaml:"child_count"`
	Status     models.Status  `json:"status" yaml:"status"`
	Input      *data.Pipeline `json:"input" yaml:"input"`
	Output     *data.Pipeline `json:"output" yaml:"output"`
}
